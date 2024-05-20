package kline

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type HuobiHistory struct {
	Pair     string   `json:"pair"`   //交易对
	Period   string   `json:"period"` //时段
	ProxyUrl *url.URL `json:"proxy"`
}

// k线数据 蜡烛图
// 1min, 5min, 15min, 30min, 60min, 4hour, 1day, 1mon, 1week, 1year
func (h *HuobiHistory) GetHuoBiHistory() error {

	per := h.Period
	if h.Period == FourHours {
		per = "4hour"
	}

	var (
		resp *http.Response
		err  error
		body []byte
		urls = fmt.Sprintf("https://api-aws.huobi.pro/market/history/kline?symbol=%s&size=2000&period=%s", h.Pair, per)
	)

	request, _ := http.NewRequest("GET", urls, nil)

	httpClient := new(http.Client)
	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyURL(h.ProxyUrl),
	}
	if resp, err = httpClient.Do(request); err != nil {
		return err
	}

	defer resp.Body.Close()

	if body, err = io.ReadAll(resp.Body); err != nil {
		return err
	}

	type Kline struct {
		Id     int64   `json:"id"`
		Open   float64 `json:"open"`
		Close  float64 `json:"close"`
		Low    float64 `json:"low"`
		High   float64 `json:"high"`
		Amount float64 `json:"amount"`
		Vol    float64 `json:"vol"`
		Count  int     `json:"count"`
	}

	type KlineData struct {
		Ch     string  `json:"ch"`
		Status string  `json:"status"`
		Ts     int64   `json:"ts"`
		Data   []Kline `json:"data"`
	}

	data := KlineData{}
	if err = json.Unmarshal(body, &data); err != nil {
		return err
	}
	if len(data.Data) < 1 {
		return errors.New("no data")
	}
	for _, item := range data.Data {
		market := &MarketQuotations{
			Id:    item.Id,
			Pair:  h.Pair,
			Open:  item.Open,
			Close: item.Close,
			High:  item.High,
			Low:   item.Low,
			Vol:   item.Vol,
		}
		//是否开启了 周线 月线
		MarketHistoryChannel <- &MarketHistory{
			Period:           h.Period,
			Pair:             h.Pair,
			MarketQuotations: market,
		}
	}
	return nil
}
