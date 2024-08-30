package main

import (
	"encoding/json"
	"fmt"
	"github.com/jksusu/kline"
	"log"
	"os"
	"strings"
)

func main() {
	//HuobiHistory()

	//SinaHistory()
	Binance()
	//Huobi()
	//arr, err := (&kline.BaiduHistory{}).GetSharesDayHistory("AAPL")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(arr)
	//Sina()
}

// 火币
func Binance() {
	c := new(kline.Binance)
	go c.NewClient().SetProxy("socks5://localhost:1080").SetPeriod([]string{kline.AMinute}).SetPairs([]string{"btcusdt"}).History()
	for {
		select {
		case p := <-kline.MarketChannel:
			log.Println(p)
			break
		case p := <-kline.RawData:
			log.Println(p)
			break
		case p := <-kline.MarketHistoryChannel:
			fmt.Println(p)
		}
	}
}

// 火币
func Huobi() {

	c := new(kline.Huobi)
	//.SetProxy("socks5://localhost:1080")
	go c.NewClient().SetProxy("socks5://localhost:1080").SetPeriod([]string{kline.AMinute}).SetPairs([]string{"btcusdt", "ethusdt"}).Start()
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
	data := ReadJsonFile("./sina/futures/global_futures.json")

	//处理成对应的格式
	var pairs []string
	for k, _ := range data {
		if len(pairs) >= 5 {
			continue
		}
		//kk := "hf_" + k
		kk := "gb_" + strings.ToLower(k)
		pairs = append(pairs, kk)
	}

	s := (&kline.Sina{}).NewClient()
	go s.SetPairs(pairs).Start()
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

func ReadJsonFile(filename string) map[string]string {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	data := string(bytes)
	m := map[string]string{}
	json.Unmarshal([]byte(data), &m)
	return m
}
