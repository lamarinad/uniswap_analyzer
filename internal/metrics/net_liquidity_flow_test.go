package metrics

import (
	"math/big"
	"testing"
)

func TestNetLiquidityFlow(t *testing.T) {
	eth := func(n int64) *big.Int { return new(big.Int).Mul(big.NewInt(n), big.NewInt(1e18)) }
	usdc := func(n int64) *big.Int { return big.NewInt(n * 1e6) }

	got := NetLiquidityFlow([]LiquidityChange{
		{Added: true, Amount0: usdc(1000), Amount1: eth(1)}, // добавили 1000 USDC + 1 ETH
		{Added: false, Amount0: usdc(500), Amount1: eth(2)}, // забрали 500 USDC + 2 ETH
		{Added: false, Amount0: usdc(0), Amount1: eth(0)},   // пустой Burn - не считается
	}, 3000)

	// USDC: 1000 - 500 = +500, ETH: 1 - 2 = -1, в долларах: 500 - 3000 = -2500
	if got.Adds != 1 || got.Removes != 1 {
		t.Fatalf("adds=%d removes=%d, want 1 and 1", got.Adds, got.Removes)
	}
	if got.NetUSDC != 500 || got.NetWETH != -1 || got.NetUSD != -2500 {
		t.Fatalf("got %+v", got)
	}
}
