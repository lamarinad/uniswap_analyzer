// Метрика 13 "Volume/TVL"
// Объем торгов за период деленный на TVL: какая доля всех денег пула прошла через обмены
// Чем больше -> тем активнее "работают" деньги поставщиков и тем больше комиссий они получают

package metrics

// VolumeTVL: volume - объем торгов в USDC (пункт 4), tvl - все деньги пула в USDC (пункт 10)
// ok = false, если TVL не больше нуля и делить не на что
func VolumeTVL(volume, tvl float64) (ratio float64, ok bool) {
	if tvl <= 0 {
		return 0, false
	}
	return volume / tvl, true
}
