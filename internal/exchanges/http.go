package exchanges

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// BidAsk - лучшие заявки в стакане биржи:
// Bid - самая высокая цена, по которой сейчас готовы купить ETH
// Ask - самая низкая цена, по которой сейчас готовы продать ETH
type BidAsk struct {
	Bid float64
	Ask float64
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

func getJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ответ %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// parseBidAsk - биржи отдают цены строками ("1801.41"), переводим в числа
func parseBidAsk(bid, ask string) (BidAsk, error) {
	b, err := strconv.ParseFloat(bid, 64)
	if err != nil {
		return BidAsk{}, fmt.Errorf("bid %q: %w", bid, err)
	}
	a, err := strconv.ParseFloat(ask, 64)
	if err != nil {
		return BidAsk{}, fmt.Errorf("ask %q: %w", ask, err)
	}
	return BidAsk{Bid: b, Ask: a}, nil
}
