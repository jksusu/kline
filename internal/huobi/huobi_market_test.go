package huobi

import (
	"fmt"
	"github.com/jksusu/kline"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	client.SetIfRowData(true).SetProxy("socks5://localhost:1080").SetPeriod([]string{"1min", "5min"}).SetSubscribePair([]string{"btcusdt"})
	go client.Start()

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
