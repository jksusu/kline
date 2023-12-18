package sina

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	"kline"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

var (
	WebSocketClient *websocket.Conn
	IfRowData       = false //是否需要原始数据
)

type SinaClient struct {
	Header http.Header
	Dialer *websocket.Dialer
	Pairs  string //交易对
}

func NewClient() *SinaClient {
	return &SinaClient{
		Header: http.Header{},
		Dialer: &websocket.Dialer{},
	}
}

// host = socks5://localhost:1080
func (h *SinaClient) SetProxy(host string) *SinaClient {
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

// 设置头
func (s *SinaClient) SetHeader(key, val string) *SinaClient {
	s.Header.Set(key, val)
	return s
}

func (s *SinaClient) SetIfRowData(b bool) *SinaClient {
	IfRowData = b
	return s
}

// 支持不同种类，多个订阅
// pairs = hf_GC,hf_SI,hf_CAD,hf_HG,hf_AU
func (h *SinaClient) SetSubscribePair(pairs string) *SinaClient {
	h.Pairs = pairs
	return h
}

func (s *SinaClient) WebsocketConnect() (conn *websocket.Conn, err error) {
	header := s.Header
	if len(header.Get("User-Agent")) == 0 {
		header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
	}
	if len(header.Get("Origin")) == 0 {
		header.Set("Origin", "https://wap.cj.sina.cn")
	}
	dialer := websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	host := fmt.Sprintf("wss://hq.sinajs.cn/wskt?list=%s", s.Pairs)
	WebSocketClient, _, err = dialer.Dial(host, header)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			if e := WebSocketClient.WriteMessage(websocket.TextMessage, []byte("ping")); e != nil {
				runtime.Goexit()
			}
			time.Sleep(20 * time.Second)
		}
	}()
	return
}

func (s *SinaClient) Start() {
	if _, err := s.WebsocketConnect(); err != nil {
		panic("sina conn fail")
	}
	for {
		_, buf, err := WebSocketClient.ReadMessage()
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
						var marketQuotations *kline.MarketQuotations
						if prefix[0] == "hf" {
							marketQuotations = DecodePreciousMetalFutures(market, pair) //贵金属
						} else if prefix[0] == "fx" {
							marketQuotations = DecodeForeignExchange(market, pair) //外汇
						}
						if marketQuotations != nil {
							marketQuotations.Pair = pair
							kline.MarketChannel <- marketQuotations
							if IfRowData {
								kline.RawData <- &map[string]interface{}{pair: market}
							}
						}
					}
				}
			}
		}
	}
}
