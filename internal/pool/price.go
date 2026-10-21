package pool

import (
	"math"
	"math/big"
)

// В этом пуле token0=USDC(6), token1=WETH(18)
// Из sqrtPriceX96 получаем цену 1 ETH в USDC
func ETHPriceUSDC(sqrtPriceX96 *big.Int) float64 {
	sqrtP := new(big.Float).SetInt(sqrtPriceX96)
	two96 := new(big.Float).SetFloat64(math.Pow(2, 96))

	ratio := new(big.Float).Quo(sqrtP, two96) // sqrt(price)
	priceToken1PerToken0 := new(big.Float).Mul(ratio, ratio)

	// raw: сколько WETH за 1 USDC -> нужна обратная + decimals
	inv := new(big.Float).Quo(big.NewFloat(1), priceToken1PerToken0)
	inv.Mul(inv, big.NewFloat(1e12)) // 10^18-6

	f, _ := inv.Float64()
	return f
}
