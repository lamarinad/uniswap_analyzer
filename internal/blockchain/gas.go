// Газ: сколько сейчас стоит отправить транзакцию
// запрос eth_feeHistory: по каждому из последних блоков узел отдает base fee и чаевые "средней" транзакции (50-й процентиль),
// а последним числом в списке base fee - base fee следующего блока, в который попадет транзакция, отправленная сейчас
// (сеть считает его заранее по заполненности последнего блока)
// готовая подсказка узла eth_maxPriorityFeePerGas у MEV Blocker почти ноль (70 wei), поэтому чаевые смотрим по истории

package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

// 5 блоков * 12 сек = последняя минута
// пункт 16 про газ сейчас, а газ меняется быстро поэтому берем короткое окно
// из 5 чисел середина (медиана) все равно не поддается одному-двум блокам со странными чаевыми
const tipBlocks = 5

type RecentFees struct {
	NextBaseFee *big.Int   // base fee следующего блока, wei
	Tips        []*big.Int // чаевые "средней" транзакции в каждом из последних tipBlocks блоков, wei
}

// GasFees - base fee следующего блока и чаевые последних tipBlocks блоков до lastBlock
func GasFees(ctx context.Context, client *ethclient.Client, lastBlock *big.Int) (RecentFees, error) {
	fh, err := client.FeeHistory(ctx, tipBlocks, lastBlock, []float64{50})
	if err != nil {
		return RecentFees{}, err
	}
	// BaseFee: tipBlocks чисел по прошедшим блокам + 1 число для следующего блока
	if len(fh.BaseFee) == 0 {
		return RecentFees{}, fmt.Errorf("eth_feeHistory: пустой список base fee")
	}
	tips := make([]*big.Int, 0, len(fh.Reward))
	for _, r := range fh.Reward {
		if len(r) > 0 {
			tips = append(tips, r[0])
		}
	}
	return RecentFees{NextBaseFee: fh.BaseFee[len(fh.BaseFee)-1], Tips: tips}, nil
}
