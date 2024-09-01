package kline

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/jksusu/kline/utils"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type BinanceHistory struct {
	Pair     string   `json:"pair"`   //交易对
	Period   string   `json:"period"` //时段
	ProxyUrl *url.URL `json:"proxy"`
}

func (h *HuobiHistory) GetBinanceHistory() error {
	var (
		resp *http.Response
		err  error
		body []byte
		urls = fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&limit=1000&interval=%s", strings.ToUpper(h.Pair), h.Period)
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
	data := string(body)

	if data != "" {
		arr := [][]interface{}{}
		json.Unmarshal([]byte(data), &arr)
		if len(arr) > 0 {
			for _, v := range arr {
				hq := &MarketHistory{
					Period: h.Period,
					Pair:   h.Pair,
					MarketQuotations: &MarketQuotations{
						Id:     utils.ConvertStringToInt64(fmt.Sprintf("%v", v[0])),
						Pair:   h.Pair,
						Period: h.Period,
						Open:   utils.ConvertStringToFloat64(fmt.Sprintf("%v", v[1])),
						Close:  utils.ConvertStringToFloat64(fmt.Sprintf("%v", v[4])),
						High:   utils.ConvertStringToFloat64(fmt.Sprintf("%v", v[2])),
						Low:    utils.ConvertStringToFloat64(fmt.Sprintf("%v", v[3])),
						Vol:    utils.ConvertStringToFloat64(fmt.Sprintf("%v", v[5])),
						Amount: utils.ConvertStringToFloat64(fmt.Sprintf("%v", v[7])),
					},
				}
				MarketHistoryChannel <- hq
			}
		}
	}
	return nil
}
