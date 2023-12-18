package main

import (
	"fmt"
	"kline"
	"kline/internal/huobi"
	"kline/internal/sina"
)

func main() {
	go Huobi()
	Sina()
}

// 火币
func Huobi() {
	//设置是否需要原始数据，设置代理，设置订阅时段，设置订阅的交易对，支持链式调用
	//如果不存在网络问题可去掉 SetProxy
	//如果不需要原始数据可去掉 SetIfRowData 原始数据在 kline.RowData 中订阅 map[string]interface{}{"交易对":"原始数据"}
	//SetSubscribePair 订阅的交易对，命名必须要符合火币的要求
	//SetPeriod 订阅的时段，建议订阅 订阅1min,其他时段可通过 kline.GetTimeModeling() 方法进行切割
	//Start 启动
	go huobi.NewClient().SetIfRowData(false).SetProxy("socks5://localhost:1080").SetPeriod([]string{"1min"}).SetSubscribePair([]string{"btcusdt"}).Start()
	for {
		select {
		case p := <-kline.MarketChannel:
			fmt.Println(p)
			break
		case p := <-kline.RawData:
			fmt.Println(p)
			break
		}
	}
}

func Sina() {
	//设置是否需要原始数据，设置代理，设置订阅时段，设置订阅的交易对，支持链式调用
	//如果不存在网络问题可去掉 SetProxy
	//如果不需要原始数据可去掉 SetIfRowData 原始数据在 kline.RowData 中订阅 map[string]interface{}{"交易对":"原始数据"}
	//SetSubscribePair 订阅的交易对，命名必须符合新浪的要求
	//Start 启动
	pair := "hf_GC,fx_saudjpy" //支持贵金属 hf 外汇 fx行情
	go sina.NewClient().SetSubscribePair(pair).Start()
	for {
		select {
		case p := <-kline.MarketChannel:
			fmt.Println(p)
			break
		case p := <-kline.RawData:
			fmt.Println(p)
			break
		}
	}
}
