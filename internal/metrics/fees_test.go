package metrics

import (
	"math"
	"math/big"
	"testing"
)

func TestPoolFees(t *testing.T) {
	eth := func(n int64) *big.Int { return new(big.Int).Mul(big.NewInt(n), big.NewInt(1e18)) }
	usdc := func(n int64) *big.Int { return big.NewInt(n * 1e6) }

	swaps := []FeeSwap{
		// купили ETH: отдали 10000 USDC, получили 5 ETH -> комиссия 10000 * 0.05% = 5 USDC
		{Amount0: usdc(10000), Amount1: new(big.Int).Neg(eth(5)), ETHPriceUSDC: 2000},
		// продали ETH: отдали 2 ETH, получили 4000 USDC -> комиссия 2 * 0.05% = 0.001 ETH = 2 USDC
		{Amount0: new(big.Int).Neg(usdc(4000)), Amount1: eth(2), ETHPriceUSDC: 2000},
	}
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

	// протокол ничего не берет: все поставщикам
	got := PoolFees(swaps, 500, 0)
	if !near(got.USDC, 5) || !near(got.WETH, 0.001) || !near(got.WETHinUSD, 2) || !near(got.TotalUSD, 7) {
		t.Fatalf("feeProtocol=0: got %+v, want USDC=5 WETH=0.001 WETHinUSD=2 TotalUSD=7", got)
	}

	// протокол берет 1/4 с USDC и 1/5 с WETH: 4 + 5*16 = 84
	got = PoolFees(swaps, 500, 84)
	if !near(got.USDC, 3.75) || !near(got.WETH, 0.0008) || !near(got.WETHinUSD, 1.6) || !near(got.TotalUSD, 5.35) {
		t.Fatalf("feeProtocol=84: got %+v, want USDC=3.75 WETH=0.0008 WETHinUSD=1.6 TotalUSD=5.35", got)
	}
}
