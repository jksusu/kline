package kline

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jksusu/kline/utils"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	baiduWSHost          = "wss://finance-ws.pae.baidu.com/"
	baiduSourcePCWeb     = "pc-web"
	baiduSnapshotProduct = "snapshot"
	baiduTickProduct     = "tick"
)

type (
	Baidu struct {
		*Client
		ProxyURL  *url.URL
		PairAlias map[string]string
		WriteLock *sync.Mutex
	}

	baiduSubscriptionItem struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Market      string `json:"market"`
		FinanceType string `json:"financeType"`
	}

	baiduSubscribeRequest struct {
		Method  string                  `json:"method"`
		Source  string                  `json:"source"`
		Product string                  `json:"product"`
		Items   []baiduSubscriptionItem `json:"items"`
	}

	baiduPingRequest struct {
		Method string `json:"method"`
		Source string `json:"source"`
	}

	baiduEnvelope struct {
		QueryID    string          `json:"queryId"`
		Data       json.RawMessage `json:"data"`
		ResultCode string          `json:"resultCode"`
	}

	baiduSnapshotData struct {
		FinanceType string `json:"financeType"`
		Code        string `json:"code"`
		Market      string `json:"market"`
		Product     string `json:"product"`
		Method      string `json:"method"`
		Cur         struct {
			Price string `json:"price"`
		} `json:"cur"`
		Point struct {
			Price           string `json:"price"`
			TotalVolume     string `json:"totalVolume"`
			TotalAmount     string `json:"totalAmount"`
			Timestamp       string `json:"timestamp"`
			RealTimeStampMs string `json:"realTimeStampMs"`
		} `json:"point"`
		Update struct {
			Time string `json:"time"`
		} `json:"update"`
		PankouInfos []baiduPankouInfo `json:"pankouinfos"`
		AskInfos    []struct {
			AskPrice  string `json:"askprice"`
			AskVolume string `json:"askvolume"`
		} `json:"askinfos"`
		BuyInfos []struct {
			BidPrice  string `json:"bidprice"`
			BidVolume string `json:"bidvolume"`
		} `json:"buyinfos"`
	}

	baiduPankouInfo struct {
		EName       string  `json:"ename"`
		OriginValue float64 `json:"originValue"`
	}
)

func (b *Baidu) NewClient() LiveMarketData {
	return &Baidu{
		Client: &Client{
			Header:           &http.Header{},
			Dialer:           &websocket.Dialer{},
			IfRowData:        false,
			WsHost:           baiduWSHost,
			LastActivityTime: time.Now().Unix(),
			MessageNumber:    0,
			ReconnectNumber:  0,
		},
		PairAlias: map[string]string{},
		WriteLock: &sync.Mutex{},
	}
}

func (b *Baidu) SetRowData(ifRow bool) LiveMarketData {
	b.IfRowData = ifRow
	return b
}

func (b *Baidu) SetProxy(proxyHost string) LiveMarketData {
	if proxyHost == "" {
		return b
	}
	proxyURL, err := url.Parse(proxyHost)
	if err != nil {
		log.Fatal(err)
	}
	if b.Dialer == nil {
		b.Dialer = &websocket.Dialer{}
	}
	b.Dialer.Proxy = http.ProxyURL(proxyURL)
	b.ProxyURL = proxyURL
	return b
}

func (b *Baidu) SetPeriod(periods []string) LiveMarketData {
	b.Period = periods
	return b
}

func (b *Baidu) SetPairs(pairs []string) LiveMarketData {
	b.Pairs = pairs
	return b
}

func (b *Baidu) SetWsHost(host string) LiveMarketData {
	b.WsHost = host
	return b
}

func (b *Baidu) SetDialer(dialer *websocket.Dialer) LiveMarketData {
	if dialer != nil {
		b.Dialer = dialer
	}
	return b
}

func (b *Baidu) WebsocketConnect() (*websocket.Conn, error) {
	if b.Dialer == nil {
		b.Dialer = &websocket.Dialer{}
	}
	b.Dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	header := b.ensureHeader()
	conn, _, err := b.Dialer.Dial(b.WsHost, *header)
	if err != nil {
		return nil, err
	}
	b.WebSocketClient = conn
	b.ReconnectNumber += 1

	if err := b.sendSubscribe(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

func (b *Baidu) Start() {
	if len(b.Pairs) == 0 {
		log.Fatal("place set pair")
	}
	if _, err := b.WebsocketConnect(); err != nil {
		log.Fatal(err)
		return
	}

	go b.pingLoop()

	for {
		_, buf, err := b.WebSocketClient.ReadMessage()
		if err != nil {
			if _, reconnectErr := b.WebsocketConnect(); reconnectErr != nil {
				time.Sleep(5 * time.Second)
				continue
			}
			log.Println("baidu reconnect connect success")
			continue
		}
		if isBaiduPongMessage(buf) {
			continue
		}
		if b.IfRowData {
			MarketRawData <- string(buf)
		}

		quotation, depth, parseErr := parseBaiduSnapshotMessage(buf)
		if parseErr != nil {
			continue
		}
		if alias := b.PairAlias[quotation.Pair]; alias != "" {
			quotation.Pair = alias
			if depth != nil {
				depth.Pair = alias
			}
		}

		MarketChannel <- quotation
		if depth != nil && (len(depth.Asks) > 0 || len(depth.Bids) > 0) {
			DepthChannel <- depth
		}
		b.MessageNumber += 1
		if time.Now().Unix()-b.LastActivityTime >= 60 {
			log.Println(fmt.Sprintf("baidu message number:%d", b.MessageNumber))
			log.Println(fmt.Sprintf("baidu reconnect number:%d", b.ReconnectNumber-1))
			b.LastActivityTime = time.Now().Unix()
		}
	}
}

func (b *Baidu) History() error {
	return errors.New("baidu history is not implemented on the LiveMarketData client")
}

func (b *Baidu) ensureHeader() *http.Header {
	if b.Header == nil {
		b.Header = &http.Header{}
	}
	if b.Header.Get("Origin") == "" {
		b.Header.Set("Origin", "https://finance.baidu.com")
	}
	if b.Header.Get("User-Agent") == "" {
		b.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")
	}
	return b.Header
}

func (b *Baidu) pingLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		b.WriteLock.Lock()
		if b.WebSocketClient != nil {
			_ = b.WebSocketClient.WriteJSON(&baiduPingRequest{
				Method: "ping",
				Source: baiduSourcePCWeb,
			})
		}
		b.WriteLock.Unlock()
	}
}

func (b *Baidu) buildSubscribeRequests() ([]*baiduSubscribeRequest, error) {
	if len(b.Pairs) == 0 {
		return nil, errors.New("baidu pairs is empty")
	}

	items := make([]baiduSubscriptionItem, 0, len(b.Pairs))
	for _, pair := range b.Pairs {
		item, err := parseBaiduPair(pair)
		if err != nil {
			return nil, err
		}
		b.PairAlias[formatBaiduPair(item)] = pair
		items = append(items, item)
	}

	products, err := b.subscriptionProducts()
	if err != nil {
		return nil, err
	}

	requests := make([]*baiduSubscribeRequest, 0, len(products))
	for _, product := range products {
		requests = append(requests, &baiduSubscribeRequest{
			Method:  "subscribe",
			Source:  baiduSourcePCWeb,
			Product: product,
			Items:   items,
		})
	}

	return requests, nil
}

func (b *Baidu) sendSubscribe() error {
	requests, err := b.buildSubscribeRequests()
	if err != nil {
		return err
	}

	b.WriteLock.Lock()
	defer b.WriteLock.Unlock()
	for _, request := range requests {
		if err := b.WebSocketClient.WriteJSON(request); err != nil {
			return err
		}
		log.Println(fmt.Sprintf("baidu subscribe success product:%s code number:%d", request.Product, len(request.Items)))
	}
	return nil
}

func (b *Baidu) subscriptionProducts() ([]string, error) {
	if len(b.Period) == 0 {
		return []string{baiduSnapshotProduct}, nil
	}

	seen := map[string]bool{}
	products := make([]string, 0, len(b.Period))
	for _, period := range b.Period {
		product := strings.ToLower(strings.TrimSpace(period))
		switch product {
		case baiduSnapshotProduct, baiduTickProduct:
			if !seen[product] {
				seen[product] = true
				products = append(products, product)
			}
		case "":
			continue
		default:
			return nil, fmt.Errorf("unsupported baidu product: %s", period)
		}
	}

	if len(products) == 0 {
		return nil, errors.New("baidu products is empty")
	}

	return products, nil
}

func parseBaiduPair(pair string) (baiduSubscriptionItem, error) {
	pair = strings.TrimSpace(pair)
	if pair == "" {
		return baiduSubscriptionItem{}, errors.New("baidu pair is empty")
	}

	parts := strings.Split(pair, ":")
	switch len(parts) {
	case 1:
		return baiduSubscriptionItem{
			Code:        parts[0],
			Name:        parts[0],
			Market:      "ab",
			FinanceType: "stock",
		}, nil
	case 3:
		return baiduSubscriptionItem{
			Market:      parts[0],
			FinanceType: parts[1],
			Code:        parts[2],
			Name:        parts[2],
		}, nil
	default:
		if len(parts) < 4 {
			return baiduSubscriptionItem{}, fmt.Errorf("invalid baidu pair: %s", pair)
		}
		return baiduSubscriptionItem{
			Market:      parts[0],
			FinanceType: parts[1],
			Code:        parts[2],
			Name:        strings.Join(parts[3:], ":"),
		}, nil
	}
}

func parseBaiduSnapshotMessage(message []byte) (*MarketQuotations, *Depth, error) {
	var envelope baiduEnvelope
	if err := json.Unmarshal(message, &envelope); err != nil {
		return nil, nil, err
	}
	if envelope.ResultCode != "" && envelope.ResultCode != "0" {
		return nil, nil, fmt.Errorf("baidu result code: %s", envelope.ResultCode)
	}

	var snapshot baiduSnapshotData
	if err := json.Unmarshal(envelope.Data, &snapshot); err != nil {
		return nil, nil, err
	}
	if snapshot.Product != baiduSnapshotProduct {
		return nil, nil, fmt.Errorf("unsupported baidu product: %s", snapshot.Product)
	}

	item := baiduSubscriptionItem{
		Code:        snapshot.Code,
		Name:        snapshot.Code,
		Market:      snapshot.Market,
		FinanceType: snapshot.FinanceType,
	}
	pair := formatBaiduPair(item)

	quotation := &MarketQuotations{
		Id:     baiduSnapshotTimestamp(snapshot),
		Pair:   pair,
		Open:   baiduPankouValue(snapshot.PankouInfos, "open"),
		Close:  baiduSnapshotClose(snapshot),
		High:   baiduPankouValue(snapshot.PankouInfos, "high"),
		Low:    baiduPankouValue(snapshot.PankouInfos, "low"),
		Vol:    baiduSnapshotVolume(snapshot),
		Amount: baiduSnapshotAmount(snapshot),
	}

	depth := &Depth{
		Pair: pair,
		Asks: baiduAskDepth(snapshot.AskInfos),
		Bids: baiduBidDepth(snapshot.BuyInfos),
	}

	return quotation, depth, nil
}

func isBaiduPongMessage(message []byte) bool {
	var envelope baiduEnvelope
	if err := json.Unmarshal(message, &envelope); err != nil {
		return false
	}

	var payload string
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		return false
	}
	return payload == "pong"
}

func formatBaiduPair(item baiduSubscriptionItem) string {
	return fmt.Sprintf("%s:%s:%s", item.Market, item.FinanceType, item.Code)
}

func baiduPankouValue(values []baiduPankouInfo, key string) float64 {
	for _, value := range values {
		if value.EName == key {
			return value.OriginValue
		}
	}
	return 0
}

func baiduSnapshotTimestamp(snapshot baiduSnapshotData) int64 {
	if ts := utils.ConvertStringToInt64(snapshot.Point.Timestamp); ts > 0 {
		return ts
	}
	return utils.ConvertStringToInt64(snapshot.Update.Time)
}

func baiduSnapshotClose(snapshot baiduSnapshotData) float64 {
	return utils.ConvertStringToFloat64(snapshot.Cur.Price)
}

func baiduSnapshotVolume(snapshot baiduSnapshotData) float64 {
	if volume := utils.ConvertStringToFloat64(snapshot.Point.TotalVolume); volume > 0 {
		return volume
	}
	return baiduPankouValue(snapshot.PankouInfos, "volume")
}

func baiduSnapshotAmount(snapshot baiduSnapshotData) float64 {
	if amount := utils.ConvertStringToFloat64(snapshot.Point.TotalAmount); amount > 0 {
		return amount
	}
	return baiduPankouValue(snapshot.PankouInfos, "amount")
}

func baiduAskDepth(values []struct {
	AskPrice  string `json:"askprice"`
	AskVolume string `json:"askvolume"`
}) []PriceVolume {
	depth := make([]PriceVolume, 0, len(values))
	for _, value := range values {
		price := utils.ConvertStringToFloat64(value.AskPrice)
		volume := utils.ConvertStringToFloat64(value.AskVolume)
		if price > 0 && volume > 0 {
			depth = append(depth, PriceVolume{Price: price, Volume: volume})
		}
	}
	return depth
}

func baiduBidDepth(values []struct {
	BidPrice  string `json:"bidprice"`
	BidVolume string `json:"bidvolume"`
}) []PriceVolume {
	depth := make([]PriceVolume, 0, len(values))
	for _, value := range values {
		price := utils.ConvertStringToFloat64(value.BidPrice)
		volume := utils.ConvertStringToFloat64(value.BidVolume)
		if price > 0 && volume > 0 {
			depth = append(depth, PriceVolume{Price: price, Volume: volume})
		}
	}
	return depth
}
