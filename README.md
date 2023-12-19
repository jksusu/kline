> ## Kline hq
> huobi sian 实时行情 websocket 接口
>
> **Thank you!**
## Install

```shell
go get github.com/jksusu/kline
```

Examples:

```go
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
	c := new(kline.Huobi).NewClient()//初始化火币
	c.SetProxy("socks5://localhost:1080")//设置代理，如果不存在网络问题则不需要
	c.SetPeriod([]string{kline.AMinute})//设置订阅时段，huobi 请到官网文档查看
	c.SetPairs([]string{"btcusdt"})//设置需要订阅的交易对
	go c.Start()//启动系统
	
	//也支持链式调用
	go (&kline.Huobi{}).NewClient().SetProxy("socks5://localhost:1080").SetPeriod([]string{kline.AMinute}).SetPairs([]string{"btcusdt"}).Start()
	
	for {
		select {
		case p := <-kline.MarketChannel:
			fmt.Println(p)
			break
		case p := <-kline.RawData:
			//原始数据，如果设置了 SetRowData
			fmt.Println(p)
			break
		}
	}
}

func Sina() {
	//新浪直接设置 SetPairs 可用
	go (&kline.Sina{}).NewClient().SetIfRowData(false).SetPairs([]string{"hf_GC", "hf_SI", "fx_susdhkd"}).Start()
	
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

```

## Warn
> 程序未做异常处理，请自己处理异常

## License

The project is licensed under the [MIT License](LICENSE).