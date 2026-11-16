package metrics

import (
	"math"
	"testing"
)

func TestExternalPriceDeviation(t *testing.T) {
	mid := MidPrice(1999, 2001)
	if mid != 2000 {
		t.Fatalf("MidPrice = %v, want 2000", mid)
	}

	// в пуле 2002, на бирже 2000 -> в пуле дороже на 0.1%
	if got := ExternalPriceDeviation(2002, mid); math.Abs(got-0.001) > 1e-12 {
		t.Fatalf("ExternalPriceDeviation = %v, want 0.001", got)
	}
}
