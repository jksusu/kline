package kline

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

/**
 * https://gushitong.baidu.com/
 * 外汇历史API，股票历史API
 */
type (
	BaiduHistory struct {
		Name string `json:"name"` //股票或者外汇名字。
	}

	baiduResultsMink struct {
		O string `json:"o"`
		C string `json:"c"`
		L string `json:"l"`
		V string `json:"v"`
		H string `json:"h"`
		D string `json:"d"`
	}

	BaiduResponse struct {
		ResultCode int64 `json:"ResultCode"`
		ResultNum  int64 `json:"ResultNum"`
		Result     struct {
			NewMarketData struct {
				Headers    []string `json:"headers"`
				Keys       []string `json:"keys"`
				MarketData string   `json:"marketData"`
			} `json:"newMarketData"`
		} `json:"Result"`
	}
)

// GetForeignExchangeMinuteHistory 获取外汇分钟线
//
//	name   名字 :HKDUSD(大写)
//	period 支持 :1 5 15 30 60 120
func (b *BaiduHistory) GetForeignExchangeMinuteHistory(name string, period int) (result []*baiduResultsMink, err error) {
	src := fmt.Sprintf("https://finance.pae.baidu.com/vapi/v1/getquotation?group=huilv_kline&ktype=min%d&code=%s&finClientType=pc", period, name)
	return b.getHistory(src)

}

// GetForeignExchangeDayHistory 外汇历史日线
//
//	name 例:USDJPY, HKDUSD 全大写
func (b *BaiduHistory) GetForeignExchangeDayHistory(name string) (result []*baiduResultsMink, err error) {
	src := fmt.Sprintf("https://finance.pae.baidu.com/vapi/v1/getquotation?group=huilv_kline&ktype=day&code=%s&finClientType=pc", name)
	return b.getHistory(src)
}

// GetSharesMinuteHistory 获取股票分钟线
//
//	name   名字 :AAPL(大写)
//	period 支持 :1 5 15 30 60 120
func (b *BaiduHistory) GetSharesMinuteHistory(name string, period int) (result []*baiduResultsMink, err error) {
	src := fmt.Sprintf("https://finance.pae.baidu.com/vapi/v1/getquotation?group=quotation_kline_us&code=%s&ktype=min%d", name, period)
	return b.getHistory(src)
}

// GetSharesDayHistory 获取股票日线
//
//	name 例:AAPL(大写)
func (b *BaiduHistory) GetSharesDayHistory(name string) (result []*baiduResultsMink, err error) {
	src := fmt.Sprintf("https://finance.pae.baidu.com/vapi/v1/getquotation?group=quotation_kline_us&code=%s&ktype=day&end_time=%s&count=1000", name, time.Now().Format("2006-01-02"))
	return b.getHistory(src)
}

// getHistory 公共方法
func (b *BaiduHistory) getHistory(src string) (result []*baiduResultsMink, err error) {
	resp, err := http.Get(src)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if body == nil {
		return result, errors.New("body is nil")
	}
	var baiduResponse *BaiduResponse
	if err = json.Unmarshal(body, &baiduResponse); err != nil {
		return
	}
	if arr := strings.Split(baiduResponse.Result.NewMarketData.MarketData, ";"); len(arr) > 0 {
		for _, v := range arr {
			if d := strings.Split(v, ","); len(d) > 6 {
				result = append(result, &baiduResultsMink{d[3], d[1], d[2], d[5], d[6], d[4]})
			}
		}
	}
	return
}
