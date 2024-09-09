package kline

import (
	"github.com/gorilla/websocket"
	"net/http"
)

// 分时
const (
	AMinute        = "1min"
	FiveMinutes    = "5min"
	FifteenMinutes = "15min"
	Minutes        = "30min"
	AnHour         = "60min"
	TwoHours       = "120min"
	FourHours      = "240min"
	ADay           = "1day"
	AWeek          = "1week"
	OneMonth       = "1mon"
	AYear          = "1year"
)

// 数据结构定义
type (
	MarketQuotations struct {
		Id     int64   `bson:"i"` //时间戳
		Period string  `json:"pd"`
		Pair   string  `json:"p"` //交易对
		Open   float64 `json:"o"` //开价
		Close  float64 `json:"c"` //关价
		High   float64 `json:"h"` //最高
		Low    float64 `json:"l"` //最低
		Vol    float64 `json:"v"` //交易量
		Amount float64 `json:"a"` //成交额
	}
	//深度数据
	Depth struct {
		Pair string        `json:"p"` //交易对
		Asks []PriceVolume `json:"a"`
		Bids []PriceVolume `json:"b"`
	}
	PriceVolume struct {
		Price  float64 `json:"p"`
		Volume float64 `json:"v"`
	}

	MarketHistory struct {
		*MarketQuotations
		Pair   string `json:"p"`
		Period string `json:"pd"`
	}

	//实时行情
	LiveMarketData interface {
		NewClient() LiveMarketData
		SetRowData(bool) LiveMarketData
		SetProxy(string) LiveMarketData
		SetPeriod([]string) LiveMarketData
		SetPairs([]string) LiveMarketData
		SetWsHost(string) LiveMarketData
		SetDialer(*websocket.Dialer) LiveMarketData
		History() error
		WebsocketConnect() (*websocket.Conn, error)
		Start()
	}

	Client struct {
		Header           *http.Header
		Dialer           *websocket.Dialer
		Period           []string //订阅的时段
		Pairs            []string //订阅的交易对
		WebSocketClient  *websocket.Conn
		IfRowData        bool //是否需要原始数据
		WsHost           string
		ReconnectNumber  int   //重连次数
		LastActivityTime int64 //最后一次接收到消息时间
		MessageNumber    int64 //总共接收消息数量
	}
)

// 通道定义
var (
	MarketChannel        = make(chan *MarketQuotations, 2048) //行情格式
	DepthChannel         = make(chan *Depth, 2048)            //深度数据
	MarketHistoryChannel = make(chan *MarketHistory, 4096)    //历史行情
	MarketRawData        = make(chan string, 2048)            //行情原始格式
)
