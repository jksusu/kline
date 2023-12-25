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
	FourHours      = "4hour"
	ADay           = "1day"
	AWeek          = "1week"
	OneMonth       = "1mon"
	AYear          = "1year"
)

// 数据结构定义
type (
	MarketQuotations struct {
		Id    int64   `bson:"id"`    //时间戳
		Pair  string  `json:"pair"`  //交易对
		Open  float64 `json:"open"`  //开价
		Close float64 `json:"close"` //关价
		High  float64 `json:"high"`  //最高
		Low   float64 `json:"low"`   //最低
		Vol   float64 `json:"vol"`   //交易量
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
		LastActivityTime int64 //上一次消息时间
		MessageNumber    int64 //总共接收消息数量
	}
)

// 通道定义
var (
	MarketChannel = make(chan *MarketQuotations, 2048)       //标准格式
	RawData       = make(chan *map[string]interface{}, 2048) //原始格式
)
