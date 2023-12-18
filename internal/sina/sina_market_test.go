package sina

import (
	"fmt"
	"kline"
	"testing"
)

func TestNewClient(t *testing.T) {
	go NewClient().SetIfRowData(true).Subscribe("hf_GC,hf_SI,hf_CAD,hf_HG,fx_susdhkd,fx_saudjpy,fx_shkdcny")

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
