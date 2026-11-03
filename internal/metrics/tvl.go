// Метрика 10 "TVL"
// Сколько всего денег лежит в пуле в долларах: USDC + ETH по текущей цене

package metrics

import "math/big"

// TVL - содержимое пула, переведённое из "сырых" чисел в монеты и доллары
type TVL struct {
	USDC      float64 // сколько USDC в пуле
	WETH      float64 // сколько ETH в пуле
	WETHinUSD float64 // сколько стоит этот ETH в USDC
	TotalUSD  float64 // всего: USDC + WETHinUSD
}

// PoolTVL: usdcRaw - баланс USDC (6 decimals), wethRaw - баланс WETH (18 decimals)
func PoolTVL(usdcRaw, wethRaw *big.Int, ethPriceUSDC float64) TVL {
	usdc := toFloat(usdcRaw, 1e6)
	weth := toFloat(wethRaw, 1e18)
	wethUSD := weth * ethPriceUSDC
	return TVL{
		USDC:      usdc,
		WETH:      weth,
		WETHinUSD: wethUSD,
		TotalUSD:  usdc + wethUSD,
	}
}

func toFloat(raw *big.Int, decimals float64) float64 {
	f := new(big.Float).SetInt(raw)
	f.Quo(f, big.NewFloat(decimals))
	out, _ := f.Float64()
	return out
}
