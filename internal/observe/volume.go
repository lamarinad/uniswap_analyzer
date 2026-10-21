package observe

import "math/big"

// VolumeUSDC суммирует [amount0] и переводит из 6 decimals в USDC
func VolumeUSDC(amount0List []*big.Int) float64 {
	total := new(big.Int)
	for _, a := range amount0List {
		total.Add(total, new(big.Int).Abs(a))
	}
	f := new(big.Float).SetInt(total)
	f.Quo(f, big.NewFloat(1e6))
	out, _ := f.Float64()
	return out
}
