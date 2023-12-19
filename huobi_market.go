package kline

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type (
	Huobi struct{ *Client }
	//订阅请求
	SubRequest struct {
		Sub string `json:"sub"`
		Id  string `json:"id"`
	}
)

// https://www.htx.com/zh-cn/opend/newApiPages/?id=628 火币websocket文档
func (c *Huobi) NewClient() LiveMarketData {
	return &Huobi{
		&Client{
			Dialer:    &websocket.Dialer{},
			IfRowData: false,
			WsHost:    "wss://api.huobi.pro/ws",
		},
	}
}
func (c *Huobi) SetRowData(ifRow bool) LiveMarketData {
	c.IfRowData = ifRow
	return c
}

// host = socks5://localhost:1080
func (c *Huobi) SetProxy(sock5 string) LiveMarketData {
	proxyUrl, err := url.Parse(sock5)
	if err != nil {
		log.Fatal(err)
	}
	dialer := &websocket.Dialer{
		Proxy: http.ProxyURL(proxyUrl),
	}
	c.Dialer = dialer
	return c
}

// 设置需要订阅的时段 []string{"1min","5min"}
// 参考  go 的变量
func (c *Huobi) SetPeriod(periodArr []string) LiveMarketData {
	c.Period = periodArr
	return c
}

func (c *Huobi) SetPairs(pairs []string) LiveMarketData {
	c.Pairs = pairs
	return c
}

func (c *Huobi) SetWsHost(host string) LiveMarketData {
	c.WsHost = host
	return c
}

func (c *Huobi) SetDialer(dialer *websocket.Dialer) LiveMarketData {
	c.Dialer = dialer
	return c
}

// 设置订阅的交易对 []string{"btcusdt,etcusdt"}
func (c *Huobi) SetSubscribePair(pair []string) LiveMarketData {
	c.Pairs = pair
	return c
}

func (c *Huobi) WebsocketConnect() (*websocket.Conn, error) {
	var err error
	c.Dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if c.WebSocketClient, _, err = c.Dialer.Dial(c.WsHost, nil); err != nil {
		return c.WebSocketClient, err
	}
	return c.WebSocketClient, nil
}

func (c *Huobi) Start() {
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
	//发送订阅
	c.SendSubscribe()

	for {
		_, buf, e := c.WebSocketClient.ReadMessage()
		if e != nil {
			c.WebsocketConnect()
			time.Sleep(5 * time.Second)
			continue
		}
		greader, err := gzip.NewReader(bytes.NewReader(buf))
		if err != nil {
			continue
		}
		greader.Close()
		buf, e = io.ReadAll(greader)
		if e != nil {
			log.Println(e.Error())
			continue
		}
		var JSONData map[string]interface{}
		err = json.Unmarshal(buf, &JSONData)
		if err != nil || JSONData == nil {
			continue
		}
		if ping, ok := JSONData["ping"]; ok {
			c.WebSocketClient.WriteJSON(map[string]interface{}{"pong": ping})
			continue
		}
		if ch, ok := JSONData["ch"]; ok {
			title := strings.Split(ch.(string), ".")
			if len(title) > 2 && title[2] == "kline" {
				if tick, tickOk := JSONData["tick"].(map[string]interface{}); tickOk {
					MarketChannel <- &MarketQuotations{
						Id:    int64(tick["id"].(float64)),
						Pair:  title[1],
						Open:  tick["open"].(float64),
						Close: tick["close"].(float64),
						High:  tick["high"].(float64),
						Low:   tick["low"].(float64),
						Vol:   tick["vol"].(float64),
					}
					//原始数据
					if c.IfRowData {
						RawData <- &map[string]interface{}{title[1]: JSONData}
					}
				}
			}
		}
	}
}

func (c *Huobi) SendSubscribe() {
	for _, pair := range c.Pairs {
		for _, p := range c.Period {
			sub := fmt.Sprintf("market.%s.kline.%s", pair, strings.TrimSpace(p))
			id := fmt.Sprintf("%s%s", pair, p)
			_ = c.WebSocketClient.WriteJSON(&SubRequest{Sub: sub, Id: id})
			time.Sleep(100 * time.Millisecond)
		}
	}
	log.Println(fmt.Sprintf("subscribe success coin number:%d", len(c.Pairs)))
}