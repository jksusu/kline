package huobi

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jksusu/kline"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	WebSocketClient *websocket.Conn
	HuobiWs         = "wss://api.huobi.pro/ws" //https://www.htx.com/zh-cn/opend/newApiPages/?id=628 火币websocket文档
	IfRowData       = false                    //是否需要原始数据
)

type (
	HuobiClient struct {
		Dialer *websocket.Dialer
		Period []string
		Pairs  []string
	}

	//订阅请求
	SubRequest struct {
		Sub string `json:"sub"`
		Id  string `json:"id"`
	}
)

// (huobi.NewClient().SetProxy("socks5://localhost:1080").SetPeriod([]string{"1min", "5min"}).SetSubscribePair([]string{"btcusdt,etcusdt"}).Start())
// 如果不存在网络问题，请去掉 SetProxy 方法
//
// (huobi.NewClient().SetPeriod([]string{"1min", "5min"}).SetSubscribePair([]string{"btcusdt,etcusdt"}).Start())
func NewClient() *HuobiClient {
	return &HuobiClient{
		Dialer: &websocket.Dialer{},
	}
}

// host = socks5://localhost:1080
func (h *HuobiClient) SetProxy(host string) *HuobiClient {
	proxyUrl, err := url.Parse(host)
	if err != nil {
		log.Fatal(err)
	}
	dialer := &websocket.Dialer{
		Proxy: http.ProxyURL(proxyUrl),
	}
	h.Dialer = dialer
	return h
}

// 设置需要订阅的时段 []string{"1min","5min"}
// 参考 kline.go 的变量
func (h *HuobiClient) SetPeriod(periodArr []string) *HuobiClient {
	h.Period = periodArr
	return h
}

// 设置订阅的交易对 []string{"btcusdt,etcusdt"}
func (h *HuobiClient) SetSubscribePair(pair []string) *HuobiClient {
	h.Pairs = pair
	return h
}

func (h *HuobiClient) SetIfRowData(ifRow bool) *HuobiClient {
	IfRowData = ifRow
	return h
}

func (h *HuobiClient) WebsocketConnect() (*websocket.Conn, error) {
	var err error
	h.Dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if WebSocketClient, _, err = h.Dialer.Dial(HuobiWs, nil); err != nil {
		return WebSocketClient, err
	}
	return WebSocketClient, nil
}

func (h *HuobiClient) Start() {
	if h.Period == nil {
		log.Fatal("place set period")
	}
	if h.Pairs == nil {
		log.Fatal("place set pair")
	}
	if _, err := h.WebsocketConnect(); err != nil {
		log.Fatal(err)
		return
	}
	//发送订阅
	h.SendSubscribe()

	for {
		_, buf, e := WebSocketClient.ReadMessage()
		if e != nil {
			h.WebsocketConnect()
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
			WebSocketClient.WriteJSON(map[string]interface{}{"pong": ping})
		}
		if ch, ok := JSONData["ch"]; ok {
			title := strings.Split(ch.(string), ".")
			if len(title) > 2 && title[2] == "kline" {
				if tick, tickOk := JSONData["tick"].(map[string]interface{}); tickOk {
					kline.MarketChannel <- &kline.MarketQuotations{
						Id:    int64(tick["id"].(float64)),
						Pair:  title[1],
						Open:  tick["open"].(float64),
						Close: tick["close"].(float64),
						High:  tick["high"].(float64),
						Low:   tick["low"].(float64),
						Vol:   tick["vol"].(float64),
					}
					//原始数据
					if IfRowData {
						kline.RawData <- &map[string]interface{}{title[1]: JSONData}
					}
				}
			}
		}
	}
}

func (h *HuobiClient) SendSubscribe() {
	for _, pair := range h.Pairs {
		for _, p := range h.Period {
			sub := "market." + pair + ".kline." + strings.TrimSpace(p)
			id := fmt.Sprintf("%s%s", pair, p)
			_ = WebSocketClient.WriteJSON(&SubRequest{Sub: sub, Id: id})
			time.Sleep(100 * time.Millisecond)
		}
	}
	log.Println(fmt.Sprintf("subscribe success coin number:%d", len(h.Pairs)))
}
