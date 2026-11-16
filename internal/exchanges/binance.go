// Binance: лучшие заявки в стакане по паре (например ETHUSDC)
// request: https://api.binance.com/api/v3/ticker/bookTicker?symbol=ETHUSDC
// response:
// {
//  "symbol": "ETHUSDC",
//  "bidPrice": "1803.24000000",
//  "bidQty": "1.26210000",
//  "askPrice": "1803.25000000",
//  "askQty": "13.96330000"
// }

package exchanges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const binanceURL = "https://api.binance.com/api/v3/ticker/bookTicker?symbol="

// https://api.binance.com/api/v3/klines?symbol=ETHUSDC&interval=1m&startTime=...&endTime=...
// ответ: [[время открытия мс, "открытие", "макс", "мин", "закрытие", ...], ...]
const binanceKlinesURL = "https://api.binance.com/api/v3/klines?symbol=%s&interval=1m&startTime=%d&endTime=%d"

func BinanceBidAsk(ctx context.Context, symbol string) (BidAsk, error) {
	var resp struct {
		BidPrice string `json:"bidPrice"`
		AskPrice string `json:"askPrice"`
	}
	if err := getJSON(ctx, binanceURL+symbol, &resp); err != nil {
		return BidAsk{}, err
	}
	return parseBidAsk(resp.BidPrice, resp.AskPrice)
}

// BinanceMinutePrices - минутные свечи пары за интервал from..to
func BinanceMinutePrices(ctx context.Context, symbol string, from, to time.Time) ([]MinutePrice, error) {
	url := fmt.Sprintf(binanceKlinesURL, symbol, from.UnixMilli(), to.UnixMilli())
	var rows [][]json.RawMessage
	if err := getJSON(ctx, url, &rows); err != nil {
		return nil, err
	}
	out := make([]MinutePrice, 0, len(rows))
	for _, row := range rows {
		if len(row) < 5 {
			return nil, fmt.Errorf("Binance: короткая свеча")
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
		return nil, fmt.Errorf("Binance: нет свечей")
	}
	return out, nil
}
