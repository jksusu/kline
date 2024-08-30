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

// 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M 币安的时间规则

// 数据结构定义
type (
	MarketQuotations struct {
		Id     int64   `bson:"id"` //时间戳
		Period string  `json:"period"`
		Pair   string  `json:"pair"`     //交易对
		Open   float64 `json:"open"`     //开价
		Close  float64 `json:"close"`    //关价
		High   float64 `json:"high"`     //最高
		Low    float64 `json:"low"`      //最低
		Vol    float64 `json:"vol"`      //交易量
		Amount float64 `json:"amount"`   //成交额
		DOpen  float64 `json:"day_open"` //今日开盘价 = 昨日收盘价
		DHigh  float64 `json:"day_high"` //今日最高
		DLow   float64 `json:"day_low"`  //今日最低
	}

	MarketHistory struct {
		*MarketQuotations
		Pair   string `json:"pair"`
		Period string `json:"period"`
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
	MarketChannel        = make(chan *MarketQuotations, 2048) //标准格式
	MarketHistoryChannel = make(chan *MarketHistory, 4096)    //历史行情
	RawData              = make(chan string, 2048)            //原始格式
)
