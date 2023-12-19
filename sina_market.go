package kline

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jksusu/kline/utils"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

type Sina struct {
	*Client
}

func (c *Sina) NewClient() LiveMarketData {
	return &Sina{
		&Client{
			Header:    &http.Header{},
			Dialer:    &websocket.Dialer{},
			IfRowData: false,
			WsHost:    "https://wap.cj.sina.cn",
		},
	}
}

// 设置项
// 也可在处理化的时候直接赋值
func (s *Sina) SetRowData(b bool) LiveMarketData {
	s.IfRowData = b
	return s
}

// host = socks5://localhost:1080
// 支持 socks5
func (s *Sina) SetProxy(host string) LiveMarketData {
	proxyUrl, err := url.Parse(host)
	if err != nil {
		log.Fatal(err)
	}
	dialer := &websocket.Dialer{
		Proxy: http.ProxyURL(proxyUrl),
	}
	s.Dialer = dialer
	return s
}
func (s *Sina) SetPeriod(periods []string) LiveMarketData {
	s.Period = periods
	return s
}
func (s *Sina) SetPairs(pairs []string) LiveMarketData {
	s.Pairs = pairs
	return s
}
func (s *Sina) SetWsHost(host string) LiveMarketData {
	s.WsHost = host
	return s
}
func (s *Sina) SetDialer(dialer *websocket.Dialer) LiveMarketData {
	return s
}
func (s *Sina) SetHeader(key, val string) LiveMarketData {
	s.Header.Set(key, val)
	return s
}

// 支持不同种类，多个订阅
// pairs = hf_GC,hf_SI,hf_CAD,hf_HG,hf_AU
func (h *Sina) SetSubscribePair(pairs []string) *Client {
	h.Pairs = pairs
	return h.Client
}
func (s *Sina) WebsocketConnect() (conn *websocket.Conn, err error) {
	header := s.Header
	if len(header.Get("User-Agent")) == 0 {
		header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
	}
	if len(header.Get("Origin")) == 0 {
		header.Set("Origin", s.WsHost)
	}
	dialer := websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	host := fmt.Sprintf("wss://hq.sinajs.cn/wskt?list=%s", s.fmtPairs())
	s.WebSocketClient, _, err = dialer.Dial(host, *header)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			if e := s.WebSocketClient.WriteMessage(websocket.TextMessage, []byte("ping")); e != nil {
				runtime.Goexit()
			}
			time.Sleep(20 * time.Second)
		}
	}()
	return
}
func (s *Sina) Start() {
	if _, err := s.WebsocketConnect(); err != nil {
		panic("sina conn fail")
	}
	for {
		_, buf, err := s.WebSocketClient.ReadMessage()
		if err != nil {
			s.WebsocketConnect() //重连
			time.Sleep(10 * time.Second)
			continue
		}
		data := strings.Split(string(buf), "\n")
		if len(data) > 0 {
			for _, item := range data {
				hf := strings.Split(item, "=")
				//hf长度2才是正确的
				if len(hf) == 2 {
					pair := hf[0] //相对行情交易对
					if len(pair) != 0 {
						market := strings.Split(hf[1], ",") //行情数据
						prefix := strings.Split(pair, "_")  //解析pair前缀
						if len(prefix) != 2 {
							continue
						}
						var marketQuotations *MarketQuotations
						if prefix[0] == "hf" {
							marketQuotations = s.DecodePreciousMetalFutures(market, pair) //贵金属
						} else if prefix[0] == "fx" {
							marketQuotations = s.DecodeForeignExchange(market, pair) //外汇
						}
						if marketQuotations != nil {
							marketQuotations.Pair = pair
							MarketChannel <- marketQuotations
							if s.IfRowData {
								RawData <- &map[string]interface{}{pair: market}
							}
						}
					}
				}
			}
		}
	}
}
func (s *Sina) DecodePreciousMetalFutures(market []string, pair string) *MarketQuotations {
	return &MarketQuotations{
		Id:    time.Now().Unix(),
		Pair:  pair,
		Open:  utils.ConvertStringToFloat64(market[8]),
		Close: utils.ConvertStringToFloat64(market[0]),
		High:  utils.ConvertStringToFloat64(market[4]),
		Low:   utils.ConvertStringToFloat64(market[5]),
		Vol:   utils.ConvertStringToFloat64(market[9]),
	}
}
func (s *Sina) DecodeForeignExchange(market []string, pair string) *MarketQuotations {
	return &MarketQuotations{
		Id:    time.Now().Unix(),
		Pair:  pair,
		Open:  utils.ConvertStringToFloat64(market[5]),
		Close: utils.ConvertStringToFloat64(market[1]),
		High:  utils.ConvertStringToFloat64(market[6]),
		Low:   utils.ConvertStringToFloat64(market[7]),
		Vol:   utils.ConvertStringToFloat64(market[11]),
	}
}

// 格式化pairs到sina需要的格式
func (s *Sina) fmtPairs() string {
	//处理pair
	str := ""
	for _, v := range s.Pairs {
		str += v + ","
	}
	if len(str) > 0 && str[len(str)-1] == ',' {
		str = str[:len(str)-1]
	}
	return str
}
