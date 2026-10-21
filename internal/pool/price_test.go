package pool

import (
	"math/big"
	"testing"
)

func TestETHPriceUSDC(t *testing.T) {
	sqrt := new(big.Int)
	sqrt.SetString("1586104266194229566155287248889292", 10)

	price := ETHPriceUSDC(sqrt)
	if price < 2000 || price > 3500 {
		t.Fatalf("unexpected price: %v", price)
	}
}
