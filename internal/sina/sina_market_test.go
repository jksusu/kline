package sina

import (
	"fmt"
	"github.com/jksusu/kline"
	"testing"
)

func TestNewClient(t *testing.T) {
	go NewClient().SetIfRowData(true).SetSubscribePair("hf_GC,hf_SI,hf_CAD,hf_HG,fx_susdhkd,fx_saudjpy,fx_shkdcny").Start()

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
