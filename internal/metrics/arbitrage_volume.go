// Метрика 19 "Arbitrage volume"
// Объем обменов, которые подтянули цену пула к цене на биржах
// Биржевую ногу сделки не видно: на цепи остается только обмен в пуле, поэтому смотрим на три признака
// 1. до обмена цены расходились больше чем на комиссию пула
// 2. направление верное: пул дороже - продают ETH в пул, пул дешевле - покупают ETH из пула
// 3. после обмена цена пула ближе к бирже, чем до него

package metrics

import (
	"math"
	"math/big"
)

// ArbSwap - один обмен, у которого есть цена пула до и после и цена биржи в ту же минуту
type ArbSwap struct {
	Amount0       *big.Int // USDC, 6 decimals: >0 купили ETH, <0 продали ETH в пул
	PriceBefore   float64
	PriceAfter    float64
	ExternalPrice float64
}

type Arbitrage struct {
	VolumeUSDC float64 // сумма размеров таких обменов
	Count      int
	Share      float64 // доля от всего объема за час
}

// ArbitrageVolume: feeRate - комиссия пула долей (0.0005 = 0.05%), totalVolume - весь объем за час в USDC
func ArbitrageVolume(swaps []ArbSwap, feeRate, totalVolume float64) Arbitrage {
	var a Arbitrage
	for _, s := range swaps {
		if !isArbitrage(s, feeRate) {
			continue
		}
		a.VolumeUSDC += toFloat(new(big.Int).Abs(s.Amount0), 1e6)
		a.Count++
	}
	if totalVolume > 0 {
		a.Share = a.VolumeUSDC / totalVolume
	}
	return a
}

func isArbitrage(s ArbSwap, feeRate float64) bool {
	if s.ExternalPrice <= 0 || s.PriceBefore <= 0 || s.Amount0 == nil {
		return false
	}
	devBefore := (s.PriceBefore - s.ExternalPrice) / s.ExternalPrice
	devAfter := (s.PriceAfter - s.ExternalPrice) / s.ExternalPrice
	if math.Abs(devBefore) <= feeRate {
		return false
	}
	// пул дороже биржи: арбитраж продает ETH в пул (amount0 < 0) и цена падает
	// пул дешевле: арбитраж покупает ETH (amount0 > 0) и цена растет
	if devBefore > 0 && s.Amount0.Sign() >= 0 {
		return false
	}
	if devBefore < 0 && s.Amount0.Sign() <= 0 {
		return false
	}
	return math.Abs(devAfter) < math.Abs(devBefore)
}
