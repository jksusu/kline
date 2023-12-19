package main

import (
	"fmt"
	"github.com/jksusu/kline"
)

func main() {
	go Huobi()
	Sina()
}

// 火币
func Huobi() {
	c := new(kline.Huobi)
	go c.NewClient().SetProxy("socks5://localhost:1080").SetPeriod([]string{kline.AMinute}).SetPairs([]string{"btcusdt"}).Start()
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
	s := (&kline.Sina{}).NewClient()
	go s.SetRowData(false).SetPairs([]string{"hf_GC", "hf_SI", "fx_susdhkd"}).Start()
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
