// Время блоков: в заголовке блока поле timestamp, секунды с 1970 года
// Пункт 19 по нему находит минутную свечу биржи, в которую попал обмен
// Запрос eth_getBlockByNumber, пачками, и только по блокам, где были обмены

package blockchain

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// BlockTimestamps возвращает время каждого блока, в секундах
func BlockTimestamps(ctx context.Context, client *ethclient.Client, blocks []uint64) (map[uint64]uint64, error) {
	out := make(map[uint64]uint64, len(blocks))
	for start := 0; start < len(blocks); start += txBatchSize {
		end := min(start+txBatchSize, len(blocks))

		results := make([]struct {
			Timestamp *hexutil.Uint64 `json:"timestamp"`
		}, end-start)
		batch := make([]rpc.BatchElem, end-start)
		for i := range batch {
			batch[i] = rpc.BatchElem{
				Method: "eth_getBlockByNumber",
				Args:   []any{hexutil.EncodeUint64(blocks[start+i]), false},
				Result: &results[i],
			}
		}
		if err := client.Client().BatchCallContext(ctx, batch); err != nil {
			return nil, err
		}
		for i, b := range batch {
			n := blocks[start+i]
			if b.Error != nil {
				return nil, fmt.Errorf("блок %d: %w", n, b.Error)
			}
			if results[i].Timestamp == nil {
				return nil, fmt.Errorf("блок %d: узел не вернул время", n)
			}
			out[n] = uint64(*results[i].Timestamp)
		}
	}
	return out, nil
}
