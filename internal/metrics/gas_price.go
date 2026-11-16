// Метрика 16 "Gas price"
// Цена газа сейчас: base fee (обязательная, сжигается) + чаевые валидатору
// и сколько примерно стоит один обмен на Uniswap в долларах

package metrics

import (
	"math/big"
	"sort"
)

// примерно столько газа тратит обычный обмен через приложение Uniswap (допущение, у каждой сделки свое)
const SwapGas = 150_000

type Gas struct {
	BaseGwei    float64 // base fee (в пункте 16 - следующего блока)
	TipGwei     float64 // середина (медиана) чаевых по последним блокам
	TotalGwei   float64 // base fee + чаевые
	SwapCostETH float64 // SwapGas * TotalGwei
	SwapCostUSD float64 // то же в долларах
}

// GasPrice: baseFee и tips в wei (1 gwei = 10^9 wei), ethPriceUSDC - цена ETH для перевода в доллары
// если tips пустой, чаевые считаются нулем
func GasPrice(baseFee *big.Int, tips []*big.Int, ethPriceUSDC float64) Gas {
	g := Gas{BaseGwei: toFloat(baseFee, 1e9), TipGwei: medianGwei(tips)}
	g.TotalGwei = g.BaseGwei + g.TipGwei
	g.SwapCostETH = SwapGas * g.TotalGwei / 1e9
	g.SwapCostUSD = g.SwapCostETH * ethPriceUSDC
	return g
}

// medianGwei - середина списка: половина чисел меньше, половина больше
// (один блок с огромными чаевыми не исказит результат, в отличие от среднего)
func medianGwei(wei []*big.Int) float64 {
	if len(wei) == 0 {
		return 0
	}
	v := make([]float64, len(wei))
	for i, w := range wei {
		v[i] = toFloat(w, 1e9)
	}
	sort.Float64s(v)
	n := len(v)
	if n%2 == 1 {
		return v[n/2]
	}
	return (v[n/2-1] + v[n/2]) / 2
}
