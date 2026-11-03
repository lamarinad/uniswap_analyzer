package metrics

import (
	"math/big"
	"testing"
)

func TestVolumeUSDC(t *testing.T) {
	// 1_000_000 + 2_000_000 = 3 USDC
	got := VolumeUSDC([]*big.Int{
		big.NewInt(1_000_000),
		big.NewInt(2_000_000),
	})
	if got != 3 {
		t.Fatalf("got %v, want 3", got)
	}
}
