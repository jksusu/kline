package kline

import "testing"

const baiduSnapshotMessage = `{
  "queryId": "2648710523556085660",
  "data": {
    "financeType": "stock",
    "code": "002541",
    "market": "ab",
    "product": "snapshot",
    "method": "subscribe",
    "cur": {
      "avgPrice": "20.06",
      "ratio": "-1.18%",
      "increase": "-0.24",
      "price": "20.12",
      "status": "down",
      "unit": "元"
    },
    "pankouinfos": [
      {"ename": "open", "originValue": 20.36},
      {"ename": "high", "originValue": 20.44},
      {"ename": "volume", "originValue": 4166200},
      {"ename": "low", "originValue": 19.86},
      {"ename": "amount", "originValue": 83579921}
    ],
    "update": {
      "time": "1778209215",
      "text": "05-08 11:00:00"
    },
    "point": {
      "price": "20.12",
      "avgPrice": "20.06",
      "totalVolume": "4166200",
      "totalAmount": "83579921",
      "timestamp": "1778209200",
      "realTimeStampMs": "1778209215000"
    },
    "askinfos": [
      {"askprice": "20.17", "askvolume": "4600"},
      {"askprice": "20.16", "askvolume": "12100"}
    ],
    "buyinfos": [
      {"bidprice": "20.11", "bidvolume": "2000"},
      {"bidprice": "20.10", "bidvolume": "900"}
    ]
  },
  "resultCode": "0"
}`

const baiduSnapshotWithoutCurPrice = `{
  "queryId": "2648710523556085660",
  "data": {
    "financeType": "stock",
    "code": "002541",
    "market": "ab",
    "product": "snapshot",
    "method": "subscribe",
    "cur": {
      "price": ""
    },
    "pankouinfos": [
      {"ename": "open", "originValue": 20.36},
      {"ename": "high", "originValue": 20.44},
      {"ename": "volume", "originValue": 4166200},
      {"ename": "low", "originValue": 19.86},
      {"ename": "amount", "originValue": 83579921}
    ],
    "point": {
      "price": "99.99",
      "totalVolume": "4166200",
      "totalAmount": "83579921",
      "timestamp": "1778209200"
    }
  },
  "resultCode": "0"
}`

func TestParseBaiduPair_DefaultsToABStock(t *testing.T) {
	item, err := parseBaiduPair("002541")
	if err != nil {
		t.Fatalf("parseBaiduPair returned error: %v", err)
	}
	if item.Code != "002541" {
		t.Fatalf("expected code 002541, got %s", item.Code)
	}
	if item.Market != "ab" {
		t.Fatalf("expected market ab, got %s", item.Market)
	}
	if item.FinanceType != "stock" {
		t.Fatalf("expected financeType stock, got %s", item.FinanceType)
	}
	if item.Name != "002541" {
		t.Fatalf("expected default name to mirror code, got %s", item.Name)
	}
}

func TestParseBaiduPair_SupportsExtendedFormat(t *testing.T) {
	item, err := parseBaiduPair("hk:stock:00700:腾讯控股")
	if err != nil {
		t.Fatalf("parseBaiduPair returned error: %v", err)
	}
	if item.Market != "hk" || item.FinanceType != "stock" || item.Code != "00700" || item.Name != "腾讯控股" {
		t.Fatalf("unexpected parsed item: %#v", item)
	}
}

func TestBuildBaiduSubscribeRequests_DefaultsToSnapshotOnly(t *testing.T) {
	client := (&Baidu{}).NewClient().(*Baidu)
	client.SetPairs([]string{"002541", "hk:stock:00700:腾讯控股"})

	requests, err := client.buildSubscribeRequests()
	if err != nil {
		t.Fatalf("buildSubscribeRequests returned error: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected 1 subscribe request, got %d", len(requests))
	}
	if requests[0].Method != "subscribe" {
		t.Fatalf("expected subscribe methods, got %#v", requests)
	}
	if requests[0].Source != "pc-web" {
		t.Fatalf("expected source pc-web, got %#v", requests)
	}
	if requests[0].Product != "snapshot" {
		t.Fatalf("expected first product snapshot, got %s", requests[0].Product)
	}
	if len(requests[0].Items) != 2 {
		t.Fatalf("expected 2 items in both requests, got %#v", requests)
	}
	if requests[0].Items[0].Code != "002541" || requests[0].Items[0].Market != "ab" || requests[0].Items[0].FinanceType != "stock" {
		t.Fatalf("unexpected first subscribe item: %#v", requests[0].Items[0])
	}
	if requests[0].Items[1].Name != "腾讯控股" {
		t.Fatalf("expected extended name to be preserved, got %#v", requests[0].Items[1])
	}
	if client.PairAlias["ab:stock:002541"] != "002541" {
		t.Fatalf("expected simple pair alias to preserve original input, got %#v", client.PairAlias)
	}
	if client.PairAlias["hk:stock:00700"] != "hk:stock:00700:腾讯控股" {
		t.Fatalf("expected extended pair alias to preserve original input, got %#v", client.PairAlias)
	}
}

func TestBuildBaiduSubscribeRequests_UsesRequestedProducts(t *testing.T) {
	client := (&Baidu{}).NewClient().(*Baidu)
	client.SetPairs([]string{"002541"})
	client.SetPeriod([]string{"tick", "snapshot"})

	requests, err := client.buildSubscribeRequests()
	if err != nil {
		t.Fatalf("buildSubscribeRequests returned error: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("expected 2 subscribe requests, got %d", len(requests))
	}
	if requests[0].Product != "tick" {
		t.Fatalf("expected first product tick, got %s", requests[0].Product)
	}
	if requests[1].Product != "snapshot" {
		t.Fatalf("expected second product snapshot, got %s", requests[1].Product)
	}
}

func TestBuildBaiduSubscribeRequests_TickOnly(t *testing.T) {
	client := (&Baidu{}).NewClient().(*Baidu)
	client.SetPairs([]string{"002541"})
	client.SetPeriod([]string{"tick"})

	requests, err := client.buildSubscribeRequests()
	if err != nil {
		t.Fatalf("buildSubscribeRequests returned error: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected 1 subscribe request, got %d", len(requests))
	}
	if requests[0].Product != "tick" {
		t.Fatalf("expected product tick, got %s", requests[0].Product)
	}
}

func TestParseBaiduSnapshotMessage_MapsMarketAndDepth(t *testing.T) {
	quotation, depth, err := parseBaiduSnapshotMessage([]byte(baiduSnapshotMessage))
	if err != nil {
		t.Fatalf("parseBaiduSnapshotMessage returned error: %v", err)
	}

	if quotation.Pair != "ab:stock:002541" {
		t.Fatalf("expected pair ab:stock:002541, got %s", quotation.Pair)
	}
	if quotation.Id != 1778209200 {
		t.Fatalf("expected timestamp 1778209200, got %d", quotation.Id)
	}
	if quotation.Open != 20.36 || quotation.Close != 20.12 || quotation.High != 20.44 || quotation.Low != 19.86 {
		t.Fatalf("unexpected quotation values: %#v", quotation)
	}
	if quotation.Vol != 4166200 || quotation.Amount != 83579921 {
		t.Fatalf("unexpected volume values: %#v", quotation)
	}
	if depth.Pair != "ab:stock:002541" {
		t.Fatalf("expected depth pair ab:stock:002541, got %s", depth.Pair)
	}
	if len(depth.Asks) != 2 || len(depth.Bids) != 2 {
		t.Fatalf("expected 2 asks and 2 bids, got %#v", depth)
	}
	if depth.Asks[0].Price != 20.17 || depth.Asks[0].Volume != 4600 {
		t.Fatalf("unexpected ask depth: %#v", depth.Asks[0])
	}
	if depth.Bids[0].Price != 20.11 || depth.Bids[0].Volume != 2000 {
		t.Fatalf("unexpected bid depth: %#v", depth.Bids[0])
	}
}

func TestParseBaiduSnapshotMessage_DoesNotFallbackClosePrice(t *testing.T) {
	quotation, _, err := parseBaiduSnapshotMessage([]byte(baiduSnapshotWithoutCurPrice))
	if err != nil {
		t.Fatalf("parseBaiduSnapshotMessage returned error: %v", err)
	}
	if quotation.Close != 0 {
		t.Fatalf("expected close to stay 0 when cur.price is missing, got %v", quotation.Close)
	}
}

func TestIsBaiduPongMessage(t *testing.T) {
	if !isBaiduPongMessage([]byte(`{"queryId":"2648710523556085660","data":"pong","resultCode":"0"}`)) {
		t.Fatal("expected pong message to be detected")
	}
	if isBaiduPongMessage([]byte(baiduSnapshotMessage)) {
		t.Fatal("did not expect snapshot message to be treated as pong")
	}
}
