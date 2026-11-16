// Транзакции: кто отправил каждую транзакцию (поле from - кошелек, который подписал ее и заплатил за газ)
// событие Swap этого не хранит: в нем sender - контракт, вызвавший пул (маршрутизатор, бот), а не человек
// запрос eth_getTransactionByHash, отправляем пачками (batch): много запросов в одном обращении к узлу

package blockchain

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// сколько запросов в одной пачке: узел принимал и 300+, но пачки поменьше надежнее
const txBatchSize = 100

// TxSenders возвращает from для каждого хэша, в том же порядке
func TxSenders(ctx context.Context, client *ethclient.Client, hashes []common.Hash) ([]common.Address, error) {
	senders := make([]common.Address, len(hashes))
	for start := 0; start < len(hashes); start += txBatchSize {
		end := min(start+txBatchSize, len(hashes))

		// из ответа узла нам нужно только поле from
		results := make([]struct {
			From *common.Address `json:"from"`
		}, end-start)
		batch := make([]rpc.BatchElem, end-start)
		for i := range batch {
			batch[i] = rpc.BatchElem{
				Method: "eth_getTransactionByHash",
				Args:   []any{hashes[start+i]},
				Result: &results[i],
			}
		}
		if err := client.Client().BatchCallContext(ctx, batch); err != nil {
			return nil, err
		}
		for i, b := range batch {
			if b.Error != nil {
				return nil, fmt.Errorf("транзакция %s: %w", hashes[start+i], b.Error)
			}
			if results[i].From == nil {
				return nil, fmt.Errorf("транзакция %s: узел не вернул from", hashes[start+i])
			}
			senders[start+i] = *results[i].From
		}
	}
	return senders, nil
}
