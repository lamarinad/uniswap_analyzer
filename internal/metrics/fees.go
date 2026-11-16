// Метрика 12 "Fees"
// Сколько комиссий получили поставщики ликвидности за период (без доли протокола Uniswap)
// Комиссия берется с той монеты, которую трейдер отдает пулу: при покупке ETH - с USDC, при продаже - с WETH

package metrics

import "math/big"

// FeeSwap - один обмен: суммы монет и цена ETH после него (для перевода комиссии в WETH в доллары)
type FeeSwap struct {
	Amount0      *big.Int // USDC, 6 decimals: >0 вошло в пул
	Amount1      *big.Int // WETH, 18 decimals: >0 вошло в пул
	ETHPriceUSDC float64
}

type Fees struct {
	USDC      float64 // комиссии поставщикам в USDC (с покупок ETH)
	WETH      float64 // комиссии поставщикам в WETH (с продаж ETH)
	WETHinUSD float64 // сколько стоят комиссии в WETH, по цене каждого обмена
	TotalUSD  float64 // всего: USDC + WETHinUSD
}

// PoolFees: feePips - ставка пула в миллионных долях, как ее возвращает fee() (500 = 0.05%)
// feeProtocol - поле из slot0: младшие 4 бита для USDC, старшие для WETH (0 = протокол не берет, N = берет 1/N)
func PoolFees(swaps []FeeSwap, feePips uint32, feeProtocol uint8) Fees {
	rate := float64(feePips) / 1e6
	usdcRate := rate * lpShare(feeProtocol%16)
	wethRate := rate * lpShare(feeProtocol>>4)
	var f Fees
	for _, s := range swaps {
		if s.Amount0.Sign() > 0 {
			f.USDC += toFloat(s.Amount0, 1e6) * usdcRate
		}
		if s.Amount1.Sign() > 0 {
			wethFee := toFloat(s.Amount1, 1e18) * wethRate
			f.WETH += wethFee
			f.WETHinUSD += wethFee * s.ETHPriceUSDC
		}
	}
	f.TotalUSD = f.USDC + f.WETHinUSD
	return f
}

// lpShare - какая часть комиссии остается поставщикам, если протокол берет 1/n
func lpShare(n uint8) float64 {
	if n == 0 {
		return 1
	}
	return 1 - 1/float64(n)
}
