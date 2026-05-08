> ## Kline hq
> huobi sina baidu 实时行情 websocket 接口
>
> **Thank you!**
## Install

```shell
go get github.com/jksusu/kline
```

Examples: 详细功能请看 [example](example)

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
		case p := <-kline.MarketRawData:
			//原始数据，如果设置了 SetRowData
			fmt.Println(p)
			break
		}
	}
}

func Sina() {
	//新浪直接设置 SetPairs 可用
	go (&kline.Sina{}).NewClient().SetRowData(true).SetPairs([]string{"hf_GC", "hf_SI", "fx_susdhkd"}).Start()
	
	for {
		select {
		case p := <-kline.MarketChannel:
			fmt.Println(p)
			break
		case p := <-kline.MarketRawData:
			fmt.Println(p)
			break
		}
	}
}

func Baidu() {
	// 默认只订阅 snapshot；如果需要 tick，传 SetPeriod([]string{"tick"}) 或两个都传
	go (&kline.Baidu{}).NewClient().SetRowData(true).SetPeriod([]string{"snapshot", "tick"}).SetPairs([]string{"002541"}).Start()

	for {
		select {
		case p := <-kline.MarketChannel:
			fmt.Println(p)
		case p := <-kline.DepthChannel:
			fmt.Println(p)
		case p := <-kline.MarketRawData:
			fmt.Println(p)
		}
	}
}
```

百度 websocket 渠道说明：

- `SetPairs([]string{"002541"})` 会默认按 `ab:stock` 订阅 A 股股票。
- 也支持扩展格式 `market:financeType:code[:name]`，例如 `hk:stock:00700:腾讯控股`。
- 不传 `SetPeriod()` 时，默认只订阅 `snapshot`。
- `SetPeriod([]string{"tick"})` 只订阅 `tick`。
- `SetPeriod([]string{"snapshot", "tick"})` 会同时订阅两个产品，按传入顺序发送订阅请求。
- `snapshot` 会结构化输出到 `MarketChannel`、`DepthChannel`。
- `tick` 目前保留在 `MarketRawData`，等字段确认后再做结构化映射。

## Warn
> 程序未做异常处理，请自己处理异常

## License

The project is licensed under the [MIT License](LICENSE).
