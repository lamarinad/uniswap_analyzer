package observe

import "math/big"

type Flow struct {
	BuyUSDC  float64 // покупка ETH (USDC вошла в пул)
	SellUSDC float64 // продажа ETH (USDC вышло из пула)
	Count    int
}

func BuySellUSDC(amount0List []*big.Int) Flow {
	var buy, sell big.Int
	for _, a := range amount0List {
		if a.Sign() > 0 {
			buy.Add(&buy, a)
		} else if a.Sign() < 0 {
			sell.Add(&sell, new(big.Int).Abs(a))
		}
	}
	toUSDC := func(v *big.Int) float64 {
		f := new(big.Float).SetInt(v)
		f.Quo(f, big.NewFloat(1e6))
		out, _ := f.Float64()
		return out
	}

	return Flow{
		BuyUSDC:  toUSDC(&buy),
		SellUSDC: toUSDC(&sell),
		Count:    len(amount0List),
	}
}

func AvgSwapUSDC(volume float64, count int) float64 {
	if count == 0 {
		return 0
	}
	return volume / float64(count)
}
