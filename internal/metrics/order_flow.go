// Метрика 9 "Order-flow imbalance"
// Перекос между покупками и продажами ETH за период
// Число от -1 до +1: +1 - были только покупки, -1 - только продажи, 0 - поровну.

package metrics

// OrderFlowImbalance = (покупки - продажи) / (покупки + продажи)
func OrderFlowImbalance(buyUSDC, sellUSDC float64) float64 {
	total := buyUSDC + sellUSDC
	if total == 0 {
		return 0
	}
	return (buyUSDC - sellUSDC) / total
}
