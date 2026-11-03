// Метрика 6 "Средний swap size"
// Cредний размер одного обмена в USDC (объём торгов, делённый на число обменов)

package metrics

func AvgSwapUSDC(volume float64, count int) float64 {
	if count == 0 {
		return 0
	}
	return volume / float64(count)
}
