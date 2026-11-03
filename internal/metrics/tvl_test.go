// Тесты к tvl.go: проверяют перевод балансов в монеты и подсчёт TVL

package metrics

import (
	"math/big"
	"testing"
)

func TestPoolTVL(t *testing.T) {
	// 1000 USDC и 2 ETH по цене 3000 -> 1000 + 6000 = 7000
	usdcRaw := big.NewInt(1000_000_000)                             // 1000 * 10^6
	wethRaw, _ := new(big.Int).SetString("2000000000000000000", 10) // 2 * 10^18

	got := PoolTVL(usdcRaw, wethRaw, 3000)
	if got.USDC != 1000 || got.WETH != 2 || got.WETHinUSD != 6000 || got.TotalUSD != 7000 {
		t.Fatalf("got %+v", got)
	}
}
