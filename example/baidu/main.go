package main

import (
	"log"

	"github.com/jksusu/kline"
)

func main() {
	Realtime()
}

func Realtime() {
	go (&kline.Baidu{}).
		NewClient().
		SetRowData(true).
		SetPeriod([]string{"snapshot"}).
		SetPairs([]string{"002541"}).
		Start()

	for {
		select {
		case quotation := <-kline.MarketChannel:
			log.Printf("market: %+v", quotation)
		}
	}
}
