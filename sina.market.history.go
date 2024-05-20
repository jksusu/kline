package kline

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jksusu/kline/utils"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// 贵金属
type SinaHistory struct {
	Pair   string `json:"pair"`   //交易对
	Period string `json:"period"` //时段
}

// 新浪分钟线返回参数结构体
type SinaMinuteResultsMink struct {
	O string `json:"o"`
	C string `json:"c"`
	L string `json:"l"`
	V string `json:"v"`
	H string `json:"h"`
	D string `json:"d"`
}

// FinanceDispatch 贵金属分类
func (*SinaHistory) FinanceDispatch(pair, period string) {
	switch period {
	case AMinute, FiveMinutes, FifteenMinutes, Minutes, AnHour, TwoHours, FourHours:
		if err := (&SinaHistory{pair, period}).GetFinanceMinuteHistory(); err != nil {
			fmt.Println(err)
		}
	case ADay, AWeek, OneMonth, AYear:
		if err := (&SinaHistory{pair, period}).GetFinanceDayHistory(); err != nil {
			fmt.Println(err)
		}
	}
}

// 贵金属
// 新浪贵金属分钟线 支持 1,15,30,60,120,240

func (f *SinaHistory) GetFinanceMinuteHistory() error {
	pair := strings.Split(f.Pair, "_")
	if len(pair) != 2 {
		return errors.New(fmt.Sprintf("pair error len <> 2:%s", f.Pair))
	}
	symbol := strings.ToUpper(pair[1]) //请求的 symbol

	var (
		resp *http.Response
		err  error
		body []byte
		url  = fmt.Sprintf("https://gu.sina.cn/ft/api/jsonp.php/var20_NG=/GlobalService.getMink?symbol=%s&type=%s", symbol, f.Period)
	)
	if resp, err = http.Get(url); err != nil {
		return err
	}
	defer resp.Body.Close()

	if body, err = io.ReadAll(resp.Body); err != nil {
		return err
	}
	content := string(body)
	reg := regexp.MustCompile("(\\[.*\\])")
	findString := reg.FindString(content)
	if len(findString) == 0 {
		return errors.New("findString is empty data len 0")
	}
	data := []*SinaMinuteResultsMink{}
	if err = json.Unmarshal([]byte(findString), &data); err != nil {
		return err
	}
	for _, item := range data {
		market := &MarketQuotations{
			Id:    utils.ParseDate(item.D, time.DateTime),
			Pair:  f.Pair,
			Open:  utils.ConvertStringToFloat64(item.O),
			Close: utils.ConvertStringToFloat64(item.C),
			High:  utils.ConvertStringToFloat64(item.H),
			Low:   utils.ConvertStringToFloat64(item.L),
			Vol:   utils.ConvertStringToFloat64(item.V),
		}
		//是否开启了 周线 月线
		MarketHistoryChannel <- &MarketHistory{
			Period:           f.Period,
			Pair:             f.Pair,
			MarketQuotations: market,
		}
	}
	return nil
}

// 贵金属日线 支持解析 日周月年

func (f *SinaHistory) GetFinanceDayHistory() error {
	//处理交易对
	pair := strings.Split(f.Pair, "_")
	if len(pair) != 2 {
		return errors.New(fmt.Sprintf("pair error len <> 2:%s", f.Pair))
	}
	symbol := strings.ToUpper(pair[1]) //请求的 symbol

	url := fmt.Sprintf("https://stock2.finance.sina.com.cn/futures/api/openapi.php/GlobalFuturesService.getGlobalFuturesDailyKLine?symbol=%s&version=7.4.0&first_opentime=true", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return errors.New("get sina finance history data fail")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("sina finance day history data io read fai %v", err))
	}

	type Data struct {
		Date   string `json:"date"`
		Open   string `json:"open"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	}
	type Status struct {
		Code int `json:"code"`
	}
	type Result struct {
		Status Status `json:"status"`
		Data   []Data `json:"data"`
	}
	type Response struct {
		Result Result `json:"result"`
	}

	var response Response
	if err = json.Unmarshal(body, &response); err != nil {
		return errors.New(fmt.Sprintf("json unmarshal fail %v", err))
	}
	//是否空
	if len(response.Result.Data) == 0 {
		return errors.New("sina finance day history data is empty")
	}

	//写入通道
	for _, data := range response.Result.Data {

		id := utils.ParseDate(data.Date, time.DateOnly)

		//解析周月年线
		if f.Period == AWeek || f.Period == OneMonth || f.Period == AYear {
			id = GetTimeModeling(f.Period, id)
		}

		market := &MarketQuotations{
			Id:    id,
			Pair:  f.Pair,
			Open:  utils.ConvertStringToFloat64(data.Open),
			Close: utils.ConvertStringToFloat64(data.Close),
			High:  utils.ConvertStringToFloat64(data.High),
			Low:   utils.ConvertStringToFloat64(data.Low),
			Vol:   utils.ConvertStringToFloat64(data.Volume),
		}
		//是否开启了 周线 月线
		MarketHistoryChannel <- &MarketHistory{
			Period:           f.Period,
			Pair:             f.Pair,
			MarketQuotations: market,
		}
	}
	return nil
}

// ForeignExchangeDispatch 外汇历史数据
func (*SinaHistory) ForeignExchangeDispatch(pair, period string) {
	switch period {
	case AMinute, FiveMinutes, FifteenMinutes, Minutes, AnHour, TwoHours, FourHours:
		if err := (&SinaHistory{pair, period}).GetForeignExchangeMinuteHistory(); err != nil {
			fmt.Println(err)
		}
	case ADay, AWeek, OneMonth, AYear:
		//日周月年
	}
}

// 获取外汇历史数据 支持分钟 1，,5，15，30，60，120，,240
// 最多返回1000条数据

func (f *SinaHistory) GetForeignExchangeMinuteHistory() error {
	var (
		resp *http.Response
		err  error
		body []byte
		url  = fmt.Sprintf("https://vip.stock.finance.sina.com.cn/forex/api/jsonp.php/=/NewForexService.getMinKline?symbol=%s&scale=%s&datalen=1000", f.Pair, f.Period)
	)
	if resp, err = http.Get(url); err != nil {
		return err
	}
	defer resp.Body.Close()

	if body, err = io.ReadAll(resp.Body); err != nil {
		return err
	}
	content := string(body)
	reg := regexp.MustCompile("(\\[.*\\])")
	findString := reg.FindString(content)
	if len(findString) == 0 {
		errors.New("findString is empty")
	}

	data := []*SinaMinuteResultsMink{}
	if err = json.Unmarshal([]byte(findString), &data); err != nil {
		return err
	}
	if len(data) < 1 {
		return errors.New("json.Unmarshal data is null")
	}
	for _, item := range data {
		market := &MarketQuotations{
			Id:    utils.ParseDate(item.D, time.DateTime),
			Pair:  f.Pair,
			Open:  utils.ConvertStringToFloat64(item.O),
			Close: utils.ConvertStringToFloat64(item.C),
			High:  utils.ConvertStringToFloat64(item.H),
			Low:   utils.ConvertStringToFloat64(item.L),
			Vol:   utils.ConvertStringToFloat64(item.V),
		}
		//是否开启了 周线 月线
		MarketHistoryChannel <- &MarketHistory{
			Period:           f.Period,
			Pair:             f.Pair,
			MarketQuotations: market,
		}
	}
	return nil
}

func (f *SinaHistory) GetForeignExchangeDayHistory() error {
	//处理交易对
	pair := strings.Split(f.Pair, "_")
	if len(pair) != 2 {
		return errors.New(fmt.Sprintf("pair error len <> 2:%s", f.Pair))
	}
	symbol := strings.ToUpper(pair[1]) //请求的 symbol

	url := fmt.Sprintf("https://stock2.finance.sina.com.cn/futures/api/openapi.php/GlobalFuturesService.getGlobalFuturesDailyKLine?symbol=%s&version=7.4.0&first_opentime=true", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return errors.New("get sina finance history data fail")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("sina finance day history data io read fai %v", err))
	}

	type Data struct {
		Date   string `json:"date"`
		Open   string `json:"open"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	}
	type Status struct {
		Code int `json:"code"`
	}
	type Result struct {
		Status Status `json:"status"`
		Data   []Data `json:"data"`
	}
	type Response struct {
		Result Result `json:"result"`
	}

	var response Response
	if err = json.Unmarshal(body, &response); err != nil {
		return errors.New(fmt.Sprintf("json unmarshal fail %v", err))
	}
	//是否空
	if len(response.Result.Data) == 0 {
		return errors.New("sina finance day history data is empty")
	}

	//写入通道
	for _, data := range response.Result.Data {

		id := utils.ParseDate(data.Date, time.DateOnly)

		//解析周月年线
		if f.Period == AWeek || f.Period == OneMonth || f.Period == AYear {
			id = GetTimeModeling(f.Period, id)
		}

		market := &MarketQuotations{
			Id:    id,
			Pair:  f.Pair,
			Open:  utils.ConvertStringToFloat64(data.Open),
			Close: utils.ConvertStringToFloat64(data.Close),
			High:  utils.ConvertStringToFloat64(data.High),
			Low:   utils.ConvertStringToFloat64(data.Low),
			Vol:   utils.ConvertStringToFloat64(data.Volume),
		}
		//是否开启了 周线 月线
		MarketHistoryChannel <- &MarketHistory{
			Period:           f.Period,
			Pair:             f.Pair,
			MarketQuotations: market,
		}
	}
	return nil
}
