// Журнал обменов пула: скачивает все события Swap за диапазон блоков
// Каждое событие - один обмен: сколько USDC вошло/вышло и какой стала цена после него
// Из этого списка считаются объем, покупки/продажи, доходность и т.д

package blockchain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// публичные RPC не отдают логи за большой диапазон блоков за один запрос
const logChunkSize = 40

var swapABI = mustABI(`[
{
	"type": "event",
	"name": "Swap",
	"inputs": [
		{"name": "sender", "type": "address", "indexed": true},
		{"name": "recipient", "type": "address", "indexed": true},
		{"name": "amount0", "type": "int256", "indexed": false},
		{"name": "amount1", "type": "int256", "indexed": false},
		{"name": "sqrtPriceX96", "type": "uint160", "indexed": false},
		{"name": "liquidity", "type": "uint128", "indexed": false},
		{"name": "tick", "type": "int24", "indexed": false}
	]
}
]`)

var swapTopic = crypto.Keccak256Hash([]byte(
	"Swap(address,address,int256,int256,uint160,uint128,int24)",
))

// Swap - один обмен в пуле
type Swap struct {
	Block        uint64
	Amount0      *big.Int // USDC: >0 вошло в пул (купили ETH), <0 вышло (продали ETH)
	SqrtPriceX96 *big.Int // цена пула сразу после обмена
}

// FetchSwaps скачивает все обмены пула в блоках from..to (включительно), по порядку
func FetchSwaps(ctx context.Context, client *ethclient.Client, pool common.Address, from, to uint64) ([]Swap, error) {
	var swaps []Swap
	for start := from; start <= to; start += logChunkSize {
		end := min(start+logChunkSize-1, to)

		logs, err := client.FilterLogs(ctx, ethereum.FilterQuery{
			FromBlock: new(big.Int).SetUint64(start),
			ToBlock:   new(big.Int).SetUint64(end),
			Addresses: []common.Address{pool},
			Topics:    [][]common.Hash{{swapTopic}},
		})
		if err != nil {
			return nil, err
		}

		for _, lg := range logs {
			ev, err := swapABI.Unpack("Swap", lg.Data)
			if err != nil {
				return nil, err
			}
			swaps = append(swaps, Swap{
				Block:        lg.BlockNumber,
				Amount0:      ev[0].(*big.Int),
				SqrtPriceX96: ev[2].(*big.Int),
			})
		}
	}
	return swaps, nil
}
