// Метрика 11 "Net liquidity flow"
// Сколько денег поставщики ликвидности добавили в пул минус сколько забрали за период
// Плюс - денег в пуле стало больше, минус - деньги уходят

package metrics

import "math/big"

// LiquidityChange - одно добавление или изъятие ликвидности, в "сырых" числах
type LiquidityChange struct {
	Added   bool     // true - добавили, false - забрали
	Amount0 *big.Int // USDC, 6 decimals
	Amount1 *big.Int // WETH, 18 decimals
}

type LiquidityFlow struct {
	Adds        int // сколько раз добавляли
	Removes     int // сколько раз забирали
	AddedUSDC   float64
	AddedWETH   float64
	RemovedUSDC float64
	RemovedWETH float64
	NetUSDC     float64 // добавили - забрали
	NetWETH     float64
	NetUSD      float64 // NetUSDC + NetWETH по цене ETH
}

func NetLiquidityFlow(changes []LiquidityChange, ethPriceUSDC float64) LiquidityFlow {
	var f LiquidityFlow
	for _, c := range changes {
		usdc := toFloat(c.Amount0, 1e6)
		weth := toFloat(c.Amount1, 1e18)
		// Burn с нулевыми суммами ничего не забирает: так владелец позиции только пересчитывает свои комиссии
		if usdc == 0 && weth == 0 {
			continue
		}
		if c.Added {
			f.Adds++
			f.AddedUSDC += usdc
			f.AddedWETH += weth
		} else {
			f.Removes++
			f.RemovedUSDC += usdc
			f.RemovedWETH += weth
		}
	}
	f.NetUSDC = f.AddedUSDC - f.RemovedUSDC
	f.NetWETH = f.AddedWETH - f.RemovedWETH
	f.NetUSD = f.NetUSDC + f.NetWETH*ethPriceUSDC
	return f
}
