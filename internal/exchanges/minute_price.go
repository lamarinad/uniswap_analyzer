// Минутная свеча биржи: середина между ценой открытия и ценой закрытия минуты
// Пункт 19 по времени блока ищет свечу, внутри которой прошел обмен

package exchanges

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// MinutePrice - одна минута: Start - начало минуты, секунды; Mid - (открытие + закрытие) / 2
type MinutePrice struct {
	Start uint64
	Mid   float64
}

// PriceAt возвращает свечу, в которую попал момент ts (секунды)
func PriceAt(candles []MinutePrice, ts uint64) (float64, bool) {
	for _, c := range candles {
		if c.Start <= ts && ts < c.Start+60 {
			return c.Mid, true
		}
	}
	return 0, false
}

// ExternalMinutePrice - средняя цена Binance и Bybit в минуту ts
// Если ответила только одна биржа, берем ее
func ExternalMinutePrice(binance, bybit []MinutePrice, ts uint64) (float64, bool) {
	b, bok := PriceAt(binance, ts)
	y, yok := PriceAt(bybit, ts)
	switch {
	case bok && yok:
		return (b + y) / 2, true
	case bok:
		return b, true
	case yok:
		return y, true
	default:
		return 0, false
	}
}

func candleMid(openRaw, closeRaw json.RawMessage) (float64, error) {
	open, err := rawString(openRaw)
	if err != nil {
		return 0, err
	}
	close, err := rawString(closeRaw)
	if err != nil {
		return 0, err
	}
	o, err := strconv.ParseFloat(open, 64)
	if err != nil {
		return 0, fmt.Errorf("open %q: %w", open, err)
	}
	c, err := strconv.ParseFloat(close, 64)
	if err != nil {
		return 0, fmt.Errorf("close %q: %w", close, err)
	}
	return (o + c) / 2, nil
}

func rawString(raw json.RawMessage) (string, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", err
	}
	return s, nil
}

func rawUnixMs(raw json.RawMessage) (uint64, error) {
	var n uint64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n / 1000, nil
	}
	s, err := rawString(raw)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return v / 1000, nil
}
