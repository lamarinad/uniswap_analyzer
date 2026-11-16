// Метрика 3 "Реализованная волатильность"
// Насколько сильно цена "тряслась" за период
// Берем цены по порядку, считаем логарифмическую доходность каждого шага
// и складываем их квадраты: большие скачки в любую сторону сильно увеличивают результат

package metrics

import "math"

// hoursPerYear нужен, чтобы перевести часовую волатильность в годовую (так ее принято сравнивать)
const hoursPerYear = 24 * 365

// StepLogReturns возвращает логарифмическую доходность каждого шага: prices[i-1] -> prices[i]
func StepLogReturns(prices []float64) []float64 {
	if len(prices) < 2 {
		return nil
	}
	out := make([]float64, 0, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		out = append(out, LogReturn(prices[i-1], prices[i]))
	}
	return out
}

// RealizedVol = sqrt(сумма квадратов доходностей) за все окно
func RealizedVol(returns []float64) float64 {
	var sum float64
	for _, r := range returns {
		sum += r * r
	}
	return math.Sqrt(sum)
}

// AnnualizeHourlyVol переводит волатильность за 1 час в годовую: умножение на sqrt(часов в году)
func AnnualizeHourlyVol(hourlyVol float64) float64 {
	return hourlyVol * math.Sqrt(hoursPerYear)
}
