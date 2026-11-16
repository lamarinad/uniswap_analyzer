// Метрика 18 "Whale volume"
// Объем крупных сделок (китов) - обменов не меньше заданного порога в USDC
// Крупные сделки сильнее всего двигают цену, и на них часто держится заметная доля комиссий

package metrics

import "math/big"

type Whale struct {
	VolumeUSDC float64 // сумма размеров крупных сделок
	Count      int     // сколько таких сделок
	Share      float64 // доля от всего объема
}

// WhaleVolume:
// amount0List - amount0 каждого обмена (USDC, 6 decimals, знак не важен),
// thresholdUSDC - порог в USDC, сделка считается крупной при размере не меньше порога
func WhaleVolume(amount0List []*big.Int, thresholdUSDC float64) Whale {
	limit, _ := new(big.Float).Mul(big.NewFloat(thresholdUSDC), big.NewFloat(1e6)).Int(nil)
	total := new(big.Int)
	whale := new(big.Int)
	var w Whale
	for _, a := range amount0List {
		abs := new(big.Int).Abs(a)
		total.Add(total, abs)
		if abs.Cmp(limit) >= 0 {
			whale.Add(whale, abs)
			w.Count++
		}
	}
	w.VolumeUSDC = toFloat(whale, 1e6)
	if all := toFloat(total, 1e6); all > 0 {
		w.Share = w.VolumeUSDC / all
	}
	return w
}
