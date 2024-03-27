package main

import (
	"fmt"
	"github.com/jksusu/kline"
	"log"
	"os"
	"strings"
)

func main() {
	//HuobiHistory()

	//SinaHistory()
	//go Huobi()
	Sina()
}

// 火币
func Huobi() {
	c := new(kline.Huobi)
	//.SetProxy("socks5://localhost:1080")
	go c.NewClient().SetPeriod([]string{kline.AMinute}).SetPairs([]string{"btcusdt", "ethusdt"}).Start()
	for {
		select {
		case p := <-kline.MarketChannel:
			log.Println(p)
			break
		case p := <-kline.RawData:
			log.Println(p)
			break
		}
	}
}

func Sina() {
	pairs := ReadFile("./gb.txt")
	s := (&kline.Sina{}).NewClient()
	go s.SetPairs(pairs).SetRowData(true).Start()
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

func SinaHistory() {
	c := (&kline.Sina{}).NewClient()
	c.SetPairs([]string{"fx_susdhkd", "hf_GC"})
	c.SetPeriod([]string{kline.AMinute, kline.FiveMinutes, kline.FifteenMinutes, kline.Minutes, kline.AnHour, kline.TwoHours, kline.FourHours, kline.ADay, kline.AWeek, kline.OneMonth, kline.AYear})
	go c.History()

	for {
		select {
		case p := <-kline.MarketHistoryChannel:
			fmt.Println(p)
			break
		}
	}
}
func HuobiHistory() {
	c := (&kline.Huobi{}).NewClient().SetProxy("socks5://localhost:1080")
	c.SetPairs([]string{"btcusdt", "ethusdt"})
	c.SetPeriod([]string{kline.AMinute, kline.FiveMinutes, kline.FifteenMinutes, kline.Minutes, kline.AnHour, kline.TwoHours, kline.FourHours, kline.ADay, kline.AWeek, kline.OneMonth, kline.AYear})
	go c.History()

	for {
		select {
		case p := <-kline.MarketHistoryChannel:
			fmt.Println(p)
			break
		}
	}
}

func ReadFile(filename string) []string {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	data := string(bytes)
	return strings.Split(data, "\n")
}
