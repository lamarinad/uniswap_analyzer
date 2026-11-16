package metrics

import (
	"math"
	"math/big"
	"testing"
)

func TestGasPrice(t *testing.T) {
	gwei := func(n int64) *big.Int { return big.NewInt(n * 1e9) }
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

	// base fee 18 gwei, чаевые 1, 2 и 100 gwei -> медиана 2 (выброс 100 не влияет)
	// итого 20 gwei * 150000 газа = 0.003 ETH, по 2000 $ = 6 $
	got := GasPrice(gwei(18), []*big.Int{gwei(100), gwei(1), gwei(2)}, 2000)
	if !near(got.BaseGwei, 18) || !near(got.TipGwei, 2) || !near(got.TotalGwei, 20) ||
		!near(got.SwapCostETH, 0.003) || !near(got.SwapCostUSD, 6) {
		t.Fatalf("got %+v, want Base=18 Tip=2 Total=20 CostETH=0.003 CostUSD=6", got)
	}

	// без чаевых: только base fee
	if got := GasPrice(gwei(18), nil, 2000); !near(got.TotalGwei, 18) {
		t.Fatalf("no tips: got %+v, want Total=18", got)
	}
}
