package observe

import (
	"math/big"
	"testing"
)

func TestBuySellUSDC(t *testing.T) {
	flow := BuySellUSDC([]*big.Int{
		big.NewInt(2_000_000),  // +2 USDC -> buy ETH
		big.NewInt(-5_000_000), // -5 USDC -> sell ETH
	})
	if flow.BuyUSDC != 2 || flow.SellUSDC != 5 {
		t.Fatalf("got buy=%v sell=%v", flow.BuyUSDC, flow.SellUSDC)
	}
}
