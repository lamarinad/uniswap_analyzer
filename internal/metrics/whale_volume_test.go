package metrics

import (
	"math"
	"math/big"
	"testing"
)

func TestWhaleVolume(t *testing.T) {
	usdc := func(n int64) *big.Int { return big.NewInt(n * 1e6) }

	// 5000 и 99999 меньше порога 100000; 150000 и продажа на 200000 - киты
	got := WhaleVolume([]*big.Int{usdc(5_000), usdc(150_000), usdc(-200_000), usdc(99_999)}, 100_000)

	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-6 }
	if got.Count != 2 || !near(got.VolumeUSDC, 350_000) || !near(got.Share, 350_000.0/454_999) {
		t.Fatalf("got %+v, want Count=2 Volume=350000 Share=%v", got, 350_000.0/454_999)
	}

	// без обменов доля не считается
	if got := WhaleVolume(nil, 100_000); got.Count != 0 || got.Share != 0 {
		t.Fatalf("no swaps: got %+v", got)
	}
}
