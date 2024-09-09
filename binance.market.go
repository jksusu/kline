package kline

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type (
	Binance struct {
		*Client
		ProxyUrl  *url.URL
		WriteLock *sync.Mutex
	}
	//订阅请求
	BinanceSubRequest struct {
		Method string   `json:"method"`
		Params []string `json:"params"`
		Id     int      `json:"id"`
	}
)

// 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M 币安的时间规则
var (
	SUBSCRIBE          = "SUBSCRIBE"          //订阅
	UNSUBSCRIBE        = "UNSUBSCRIBE"        //取消
	LIST_SUBSCRIPTIONS = "LIST_SUBSCRIPTIONS" //获取已经订阅的信息流
	subId              = 1
	BinancePeriodMap   = map[string]string{
		AMinute:        "1m",
		FiveMinutes:    "5m",
		FifteenMinutes: "15m",
		Minutes:        "30m",
		AnHour:         "1h",
		TwoHours:       "2h",
		FourHours:      "4h",
		ADay:           "1d",
		AWeek:          "1w",
		OneMonth:       "1M",
	}
	BinancePeriodMapValKey map[string]string
)

// https://developers.binance.com/docs/zh-CN/binance-spot-api-docs/web-socket-streams
func (c *Binance) NewClient() LiveMarketData {
	BinancePeriodMapValKey = c.mapReverseValKey(BinancePeriodMap) //反向映射

	return &Binance{
		Client: &Client{
			Dialer:           &websocket.Dialer{},
			IfRowData:        false,
			WsHost:           "wss://stream.binance.com:9443/ws",
			LastActivityTime: time.Now().Unix(),
			MessageNumber:    0,
			ReconnectNumber:  0,
		},
		WriteLock: &sync.Mutex{},
	}
}

func (c *Binance) mapReverseValKey(maps map[string]string) map[string]string {
	var m = map[string]string{}
	for key, value := range maps {
		m[value] = key
	}
	return m
}

func (c *Binance) SetRowData(ifRow bool) LiveMarketData {
	c.IfRowData = ifRow
	return c
}

// host = socks5://localhost:1080
func (c *Binance) SetProxy(sock5 string) LiveMarketData {
	if sock5 == "" {
		log.Println("sock5 is null")
		return c
	}
	proxyUrl, err := url.Parse(sock5)
	if err != nil {
		log.Fatal(err)
	}
	dialer := &websocket.Dialer{
		Proxy: http.ProxyURL(proxyUrl),
	}
	c.Dialer = dialer
	c.ProxyUrl = proxyUrl
	return c
}

// 设置需要订阅的时段 []string{"1min","5min"}
// 参考  go 的变量
func (c *Binance) SetPeriod(periodArr []string) LiveMarketData {
	c.Period = periodArr
	return c
}
func (c *Binance) SetPairs(pairs []string) LiveMarketData {
	c.Pairs = pairs
	return c
}

func (c *Binance) SetWsHost(host string) LiveMarketData {
	c.WsHost = host
	return c
}

func (c *Binance) SetDialer(dialer *websocket.Dialer) LiveMarketData {
	c.Dialer = dialer
	return c
}

// 设置订阅的交易对 []string{"btcusdt,etcusdt"}
func (c *Binance) SetSubscribePair(pair []string) LiveMarketData {
	c.Pairs = pair
	return c
}

func (c *Binance) WebsocketConnect() (*websocket.Conn, error) {
	var err error
	c.Dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if c.WebSocketClient, _, err = c.Dialer.Dial(c.WsHost, nil); err != nil {
		return c.WebSocketClient, err
	}
	c.ReconnectNumber += 1

	c.SendSubscribe() //发送订阅 #bug防止重连后无数据

	return c.WebSocketClient, nil
}

func (c *Binance) Start() {
	if c.Period == nil {
		log.Fatal("place set period")
	}
	if c.Pairs == nil {
		log.Fatal("place set pair")
	}
	if _, err := c.WebsocketConnect(); err != nil {
		log.Fatal(err)
		return
	}

	for {
		_, buf, e := c.WebSocketClient.ReadMessage()
		if e != nil {
			if _, err := c.WebsocketConnect(); err != nil {
				time.Sleep(5 * time.Second)
				continue
			}
			log.Println("Binance reconnect connect success")
		}
		data := string(buf)
		if data == "ping" {
			c.WebSocketClient.WriteMessage(websocket.TextMessage, []byte("pong"))
			continue
		}
		//订阅成功的列表
		if resultArr := gjson.Get(data, "result").Array(); len(resultArr) > 0 {
			fmt.Println("行情订阅成功")
			fmt.Println(resultArr)
			continue
		}
		//归集行情事件
		event := gjson.Get(data, "e").String()
		if event == "kline" {
			MarketChannel <- &MarketQuotations{
				Id:     gjson.Get(data, "k.T").Int(),
				Period: BinancePeriodMapValKey[gjson.Get(data, "k.i").String()],
				Pair:   strings.ToLower(gjson.Get(data, "k.s").String()),
				Open:   gjson.Get(data, "k.o").Float(),
				Close:  gjson.Get(data, "k.c").Float(),
				High:   gjson.Get(data, "k.h").Float(),
				Low:    gjson.Get(data, "k.l").Float(),
				Vol:    gjson.Get(data, "k.v").Float(),
				Amount: gjson.Get(data, "k.q").Float(),
			}
			//原始数据
			if c.IfRowData {
				MarketRawData <- data
			}
		} else if event == "depthUpdate" {
			pair := strings.ToLower(gjson.Get(data, "s").String())
			c.Depth(gjson.Get(data, "b").Array(), pair)
			c.Depth(gjson.Get(data, "a").Array(), pair)
		}

		c.MessageNumber += 1
		if time.Now().Unix()-c.LastActivityTime >= 60 {
			log.Println(fmt.Sprintf("Binance message number:%d", c.MessageNumber))
			log.Println(fmt.Sprintf("Binance reconnect number:%d", c.ReconnectNumber-1))
			c.LastActivityTime = time.Now().Unix()
		}
	}
}

// 深度数据
func (h *Binance) Depth(arr []gjson.Result, pair string) {
	if len(arr) > 0 {
		var priceVolumes []PriceVolume
		for _, v := range arr {
			price := v.Array()[0].Float()
			volume := v.Array()[1].Float()
			if price > 0 && volume > 0 {
				priceVolumes = append(priceVolumes, PriceVolume{price, volume})
			}
		}
		if len(priceVolumes) > 0 {
			DepthChannel <- &Depth{
				Pair: pair,
				Asks: priceVolumes,
				Bids: priceVolumes,
			}
		}
	}
}

func (h *Binance) History() error {
	//解析设置的时段
	if len(h.Period) == 0 {
		return errors.New("binance period is empty")
	}
	if len(h.Pairs) == 0 {
		return errors.New("binance pairs is empty")
	}

	for _, pair := range h.Pairs {
		for _, period := range h.Period {
			if err := (&HuobiHistory{pair, BinancePeriodMap[period], h.ProxyUrl}).GetBinanceHistory(); err != nil {
				fmt.Println(err)
			}

			time.Sleep(5 * time.Second)
		}
	}
	return nil
}

func (c *Binance) SendSubscribe() {
	//处理 pair
	var (
		markeyPairs []string
		depthPairs  []string
	)
	for _, pair := range c.Pairs {
		depthPairs = append(depthPairs, fmt.Sprintf("%s@depth@100ms", pair))
		for _, p := range c.Period {
			markeyPairs = append(markeyPairs, fmt.Sprintf("%s@kline_%s", pair, BinancePeriodMap[p]))
		}
	}
	_ = c.WebSocketClient.WriteJSON(&BinanceSubRequest{
		Method: SUBSCRIBE,
		Params: depthPairs,
		Id:     subId,
	})
	_ = c.WebSocketClient.WriteJSON(&BinanceSubRequest{
		Method: SUBSCRIBE,
		Params: markeyPairs,
		Id:     subId,
	})
	// 获取订阅的结果
	c.WebSocketClient.WriteJSON(&BinanceSubRequest{
		Method: LIST_SUBSCRIPTIONS,
		Id:     subId,
	})

	log.Println(fmt.Sprintf("subscribe success coin number:%d", len(c.Pairs)))
}

// 设置深度数据订阅
func (c *Binance) SendSubDepth() {

}
