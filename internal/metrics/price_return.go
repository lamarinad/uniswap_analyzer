// Метрика 2 "Доходность цены"
// На сколько процентов изменилась цена за период
// Две формы: обычная (наглядная в процентах) и логарифмическая (ее можно складывать по периодам, понадобится для волатильности)

package metrics

import "math"

// SimpleReturn: во сколько раз изменилась цена, минус 1. 0.01 = +1%
func SimpleReturn(from, to float64) float64 {
	if from == 0 {
		return 0
	}
	return to/from - 1
}

// LogReturn = ln(to/from). Логарифмические доходности можно складывать по периодам
func LogReturn(from, to float64) float64 {
	if from <= 0 || to <= 0 {
		return 0
	}
	return math.Log(to / from)
}
