// Метрика 20 "MEV activity"
// Сэндвич: в одном блоке один и тот же адрес меняет монеты до чужой сделки и сразу после нее

package metrics

import (
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
)

// MEVSwap - один обмен в пуле и кошелек, который отправил транзакцию
type MEVSwap struct {
	Block   uint64
	TxIndex uint
	TxHash  common.Hash
	From    common.Address
	Amount0 *big.Int // USDC: >0 купили ETH, <0 продали ETH в пул
}

type MEV struct {
	Sandwiches int     // сколько таких связок
	BotUSDC    float64 // объем двух сделок бота
	VictimUSDC float64 // объем чужих сделок между ними
}

type mevLeg struct {
	from    common.Address
	amount  *big.Int
	txIndex uint
}

// MEVActivity ищет сэндвичи в списке обменов
func MEVActivity(swaps []MEVSwap) MEV {
	ordered := append([]MEVSwap(nil), swaps...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Block != ordered[j].Block {
			return ordered[i].Block < ordered[j].Block
		}
		return ordered[i].TxIndex < ordered[j].TxIndex
	})

	byBlock := map[uint64][]mevLeg{}
	for _, s := range ordered {
		if s.Amount0 == nil || s.Amount0.Sign() == 0 {
			continue
		}
		legs := byBlock[s.Block]
		// несколько обменов одной транзакции складываем в одну сделку
		if n := len(legs); n > 0 && legs[n-1].from == s.From && legs[n-1].txIndex == s.TxIndex {
			legs[n-1].amount.Add(legs[n-1].amount, s.Amount0)
			byBlock[s.Block] = legs
			continue
		}
		byBlock[s.Block] = append(legs, mevLeg{from: s.From, amount: new(big.Int).Set(s.Amount0), txIndex: s.TxIndex})
	}

	var m MEV
	for _, legs := range byBlock {
		sort.Slice(legs, func(i, j int) bool { return legs[i].txIndex < legs[j].txIndex })
		for j := range legs {
			back := legs[j]
			for k := j - 1; k >= 0; k-- {
				front := legs[k]
				if front.from != back.from || front.amount.Sign()*back.amount.Sign() >= 0 {
					continue
				}
				var victims []*big.Int
				for _, mid := range legs[k+1 : j] {
					if mid.from != front.from && mid.amount.Sign()*front.amount.Sign() > 0 {
						victims = append(victims, mid.amount)
					}
				}
				if len(victims) == 0 {
					continue
				}
				m.Sandwiches++
				m.BotUSDC += usdcAbs(front.amount) + usdcAbs(back.amount)
				for _, v := range victims {
					m.VictimUSDC += usdcAbs(v)
				}
				break
			}
		}
	}
	return m
}

func usdcAbs(raw *big.Int) float64 {
	return toFloat(new(big.Int).Abs(raw), 1e6)
}
