// Метрика 14 "Price impact"
// Насколько сдвинется цена пула, если прямо сейчас совершить сделку заданного размера,
// и по какой средней цене в итоге выйдет 1 ETH (с учетом комиссии пула)

package metrics

import "math/big"

// ImpactTrade - одна проверочная сделка: сколько USDC и WETH в ней участвует и цена пула после нее
// при покупке ETH: USDC - отдали, WETH - получили; при продаже наоборот
type ImpactTrade struct {
	USDCRaw    *big.Int // 6 decimals
	WETHRaw    *big.Int // 18 decimals
	PriceAfter float64  // цена ETH в пуле после сделки, USDC
}

type Impact struct {
	ETH         float64 // сколько ETH в сделке
	PriceChange float64 // сдвиг цены пула: доля, +0.001 = +0.1%
	AvgPrice    float64 // средняя цена 1 ETH в сделке, USDC
}

// PriceImpact: priceBefore - цена ETH в пуле до сделки (USDC)
func PriceImpact(priceBefore float64, t ImpactTrade) Impact {
	eth := toFloat(t.WETHRaw, 1e18)
	var avg float64
	if eth > 0 {
		avg = toFloat(t.USDCRaw, 1e6) / eth
	}
	return Impact{
		ETH:         eth,
		PriceChange: SimpleReturn(priceBefore, t.PriceAfter),
		AvgPrice:    avg,
	}
}
