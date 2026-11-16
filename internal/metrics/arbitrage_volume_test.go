package metrics

import (
	"math"
	"math/big"
	"testing"
)

func TestArbitrageVolume(t *testing.T) {
	usdc := func(n int64) *big.Int { return big.NewInt(n * 1e6) }
	swaps := []ArbSwap{
		// пул 2002, биржа 2000: продажа ведет цену к бирже (2001) - арбитраж, 100 USDC
		{Amount0: usdc(-100), PriceBefore: 2002, PriceAfter: 2001, ExternalPrice: 2000},
		// та же дорогая цена, но покупка уводит ее еще дальше - не арбитраж
		{Amount0: usdc(50), PriceBefore: 2002, PriceAfter: 2003, ExternalPrice: 2000},
		// расхождение 0.02% меньше комиссии 0.05% - не арбитраж
		{Amount0: usdc(-10), PriceBefore: 2000.4, PriceAfter: 2000.1, ExternalPrice: 2000},
		// пул 1990 дешевле биржи 2000: покупка поднимает цену к бирже - арбитраж, 80 USDC
		{Amount0: usdc(80), PriceBefore: 1990, PriceAfter: 1995, ExternalPrice: 2000},
	}

	got := ArbitrageVolume(swaps, 0.0005, 240)
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
	if got.Count != 2 || !near(got.VolumeUSDC, 180) || !near(got.Share, 180.0/240) {
		t.Fatalf("got %+v, want Count=2 Volume=180 Share=%v", got, 180.0/240)
	}
}
