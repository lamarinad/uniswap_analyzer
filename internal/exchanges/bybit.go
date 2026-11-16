// Bybit: лучшие заявки в стакане по паре на споте (обычная покупка монет, не фьючерсы)
// request: https://api.bybit.com/v5/market/tickers?category=spot&symbol=ETHUSDC
// response:
// {
// 	"retCode": 0,
// 	"retMsg": "OK",
// 	"result": {
// 	  "category": "spot",
// 	  "list": [
// 		{
// 		  "symbol": "ETHUSDC",
// 		  "bid1Price": "1800.82",
// 		  "bid1Size": "0.35913",
// 		  "ask1Price": "1800.83",
// 		  "ask1Size": "0.04",
// 		  "lastPrice": "1801.37",
// 		  "prevPrice24h": "1820.44",
// 		  "price24hPcnt": "-0.0105",
// 		  "highPrice24h": "1842.34",
// 		  "lowPrice24h": "1785.66",
// 		  "turnover24h": "6300882.6549939",
// 		  "volume24h": "3476.36639",
// 		  "usdIndexPrice": "1800.536947"
// 		}
// 	  ]
// 	},
// 	"retExtInfo": {
// 	},
// 	"time": 1794838124585
// }

// retCode = 0 - успешно, иначе в retMsg причина ошибки

package exchanges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const bybitURL = "https://api.bybit.com/v5/market/tickers?category=spot&symbol="

// https://api.bybit.com/v5/market/kline?category=spot&symbol=ETHUSDC&interval=1&start=...&end=...
// list: ["время открытия мс", "открытие", "макс", "мин", "закрытие", ...], новые свечи в начале списка
const bybitKlinesURL = "https://api.bybit.com/v5/market/kline?category=spot&symbol=%s&interval=1&start=%d&end=%d"

func BybitBidAsk(ctx context.Context, symbol string) (BidAsk, error) {
	var resp struct {
		RetCode int    `json:"retCode"`
		RetMsg  string `json:"retMsg"`
		Result  struct {
			List []struct {
				Bid1Price string `json:"bid1Price"`
				Ask1Price string `json:"ask1Price"`
			} `json:"list"`
		} `json:"result"`
	}
	if err := getJSON(ctx, bybitURL+symbol, &resp); err != nil {
		return BidAsk{}, err
	}
	if resp.RetCode != 0 {
		return BidAsk{}, fmt.Errorf("Bybit: %s", resp.RetMsg)
	}
	if len(resp.Result.List) == 0 {
		return BidAsk{}, fmt.Errorf("Bybit: нет пары %s", symbol)
	}
	t := resp.Result.List[0]
	return parseBidAsk(t.Bid1Price, t.Ask1Price)
}

// BybitMinutePrices - минутные свечи пары на споте за интервал from..to
func BybitMinutePrices(ctx context.Context, symbol string, from, to time.Time) ([]MinutePrice, error) {
	url := fmt.Sprintf(bybitKlinesURL, symbol, from.UnixMilli(), to.UnixMilli())
	var resp struct {
		RetCode int    `json:"retCode"`
		RetMsg  string `json:"retMsg"`
		Result  struct {
			List [][]json.RawMessage `json:"list"`
		} `json:"result"`
	}
	if err := getJSON(ctx, url, &resp); err != nil {
		return nil, err
	}
	if resp.RetCode != 0 {
		return nil, fmt.Errorf("Bybit: %s", resp.RetMsg)
	}
	out := make([]MinutePrice, 0, len(resp.Result.List))
	for _, row := range resp.Result.List {
		if len(row) < 5 {
			return nil, fmt.Errorf("Bybit: короткая свеча")
		}
		start, err := rawUnixMs(row[0])
		if err != nil {
			return nil, err
		}
		mid, err := candleMid(row[1], row[4])
		if err != nil {
			return nil, err
		}
		out = append(out, MinutePrice{Start: start, Mid: mid})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("Bybit: нет свечей")
	}
	return out, nil
}
