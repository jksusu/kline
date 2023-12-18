package sina

import (
	"encoding/json"
	"fmt"
	"github.com/jksusu/kline"
	"io"
	"net/http"
)

func History() {

}

// 贵金属日历史行情
func GetHfDayHistory(pair string) (market []kline.MarketQuotations, err error) {
	src := fmt.Sprintf("https://stock2.finance.sina.com.cn/futures/api/openapi.php/GlobalFuturesService.getGlobalFuturesDailyKLine?%s&&version=7.4.4&&first_opentime=true", pair)
	resp, err := http.Get(src)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if err = json.Unmarshal(body, &market); err != nil {
		return
	}
	return
}
