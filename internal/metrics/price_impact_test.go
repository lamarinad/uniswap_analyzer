package metrics

import (
	"math"
	"math/big"
	"testing"
)

func TestPriceImpact(t *testing.T) {
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

	// купили ETH: отдали 10000 USDC, получили 4.99 ETH, цена 2000 -> 2002
	got := PriceImpact(2000, ImpactTrade{
		USDCRaw:    big.NewInt(10_000 * 1e6),
		WETHRaw:    new(big.Int).Mul(big.NewInt(499), big.NewInt(1e16)),
		PriceAfter: 2002,
	})
	if !near(got.ETH, 4.99) || !near(got.PriceChange, 0.001) || !near(got.AvgPrice, 10000/4.99) {
		t.Fatalf("got %+v, want ETH=4.99 PriceChange=0.001 AvgPrice=%v", got, 10000/4.99)
	}
}
