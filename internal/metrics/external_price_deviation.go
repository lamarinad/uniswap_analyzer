// Метрика 15 "Отклонение от внешней цены"
// Насколько цена ETH в нашем пуле отличается от цены на бирже прямо сейчас:
// плюс - в нашем пуле дороже, минус - в нашем пуле дешевле

package metrics

// MidPrice - середина между лучшей заявкой на покупку (bid) и на продажу (ask)
func MidPrice(bid, ask float64) float64 {
	return (bid + ask) / 2
}

// ExternalPriceDeviation - отклонение цены пула от внешней цены, доля: +0.001 = +0.1%
func ExternalPriceDeviation(poolPrice, externalPrice float64) float64 {
	return SimpleReturn(externalPrice, poolPrice)
}
