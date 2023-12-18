package kline

// 数据结构定义
type MarketQuotations struct {
	Id    int64   `bson:"id"`
	Pair  string  `json:"pair"`
	Open  float64 `json:"open"`
	Close float64 `json:"close"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Vol   float64 `json:"vol"`
}

// 分时
var (
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

// 通道定义
var (
	MarketChannel = make(chan *MarketQuotations, 2048)       //标准格式
	RawData       = make(chan *map[string]interface{}, 2048) //原始格式
)
