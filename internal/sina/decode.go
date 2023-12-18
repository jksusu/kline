package sina

import (
	"kline"
	"kline/utils"
	"time"
)

func DecodePreciousMetalFutures(market []string, pair string) *kline.MarketQuotations {
	return &kline.MarketQuotations{
		Id:    time.Now().Unix(),
		Pair:  pair,
		Open:  utils.ConvertStringToFloat64(market[8]),
		Close: utils.ConvertStringToFloat64(market[0]),
		High:  utils.ConvertStringToFloat64(market[4]),
		Low:   utils.ConvertStringToFloat64(market[5]),
		Vol:   utils.ConvertStringToFloat64(market[9]),
	}
}

func DecodeForeignExchange(market []string, pair string) *kline.MarketQuotations {
	return &kline.MarketQuotations{
		Id:    time.Now().Unix(),
		Pair:  pair,
		Open:  utils.ConvertStringToFloat64(market[5]),
		Close: utils.ConvertStringToFloat64(market[1]),
		High:  utils.ConvertStringToFloat64(market[6]),
		Low:   utils.ConvertStringToFloat64(market[7]),
		Vol:   utils.ConvertStringToFloat64(market[11]),
	}
}
