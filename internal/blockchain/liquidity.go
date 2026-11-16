// Журнал ликвидности пула: скачивает события Mint (деньги добавили в пул) и Burn (деньги забрали из пула)
// за диапазон блоков. Каждое событие - сколько USDC (amount0) и WETH (amount1) добавили или забрали

package blockchain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Mint - деньги добавили в пул:
// - sender - кто вызвал добавление (обычно контракт-менеджер позиций Uniswap, а не кошелек человека)
// - owner - на кого записана позиция (indexed: лежит в Topics, по нему можно фильтровать)
// - tickLower - нижняя граница ценового диапазона позиции, номер тика (indexed)
// - tickUpper - верхняя граница диапазона (indexed)
// - amount - размер доли во внутренних единицах пула (ликвидность L), это не деньги
// - amount0 - сколько USDC положили, 6 знаков после запятой
// - amount1 - сколько WETH положили, 18 знаков после запятой
//
// Burn - деньги забрали из пула (поля те же, только без sender):
// - owner - чья позиция (indexed)
// - tickLower - нижняя граница диапазона (indexed)
// - tickUpper - верхняя граница диапазона (indexed)
// - amount - сколько доли закрыли, во внутренних единицах пула
// - amount0 - сколько USDC забрали, 6 знаков после запятой
// - amount1 - сколько WETH забрали, 18 знаков после запятой
//
// indexed: true - поле в Topics записи журнала, indexed: false - поле в Data.
// Unpack читает только Data по порядку, поэтому для Mint amount0/amount1 = ev[2]/ev[3], для Burn = ev[1]/ev[2]

/* {
	"type": "event",
	"name": "Mint",
	"inputs": [
		{"name": "sender", "type": "address", "indexed": false}, // кто вызвал добавление, те какой адрес нажал «добавить»
		{"name": "owner", "type": "address", "indexed": true}, // чья позиция, на кого записана доля
		{"name": "tickLower", "type": "int24", "indexed": true}, // нижняя граница ценового диапазона позиции
		{"name": "tickUpper", "type": "int24", "indexed": true}, // верхняя граница диапазона
		{"name": "amount", "type": "uint128", "indexed": false}, // размер доли во внутренних единицах пула («ликвидность»)
		{"name": "amount0", "type": "uint256", "indexed": false}, // сколько USDC положили (token0)
		{"name": "amount1", "type": "uint256", "indexed": false} // сколько WETH положили (token1)
	]
}*/

/*{
	"type": "event",
	"name": "Burn",
	"inputs": [
		{"name": "owner", "type": "address", "indexed": true}, // чья позиция
		{"name": "tickLower", "type": "int24", "indexed": true}, // нижняя граница
		{"name": "tickUpper", "type": "int24", "indexed": true}, // верхняя граница
		{"name": "amount", "type": "uint128", "indexed": false}, // сколько доли закрыли
		{"name": "amount0", "type": "uint256", "indexed": false}, // сколько USDC забрали (token0)
		{"name": "amount1", "type": "uint256", "indexed": false} // сколько WETH забрали (token1)
	]
}*/

var liquidityABI = mustABI(`[
{
	"type": "event",
	"name": "Mint",
	"inputs": [
		{"name": "sender", "type": "address", "indexed": false},
		{"name": "owner", "type": "address", "indexed": true},
		{"name": "tickLower", "type": "int24", "indexed": true},
		{"name": "tickUpper", "type": "int24", "indexed": true},
		{"name": "amount", "type": "uint128", "indexed": false},
		{"name": "amount0", "type": "uint256", "indexed": false},
		{"name": "amount1", "type": "uint256", "indexed": false}
	]
},
{
	"type": "event",
	"name": "Burn",
	"inputs": [
		{"name": "owner", "type": "address", "indexed": true},
		{"name": "tickLower", "type": "int24", "indexed": true},
		{"name": "tickUpper", "type": "int24", "indexed": true},
		{"name": "amount", "type": "uint128", "indexed": false},
		{"name": "amount0", "type": "uint256", "indexed": false},
		{"name": "amount1", "type": "uint256", "indexed": false}
	]
}
]`)

var mintTopic = crypto.Keccak256Hash([]byte(
	"Mint(address,address,int24,int24,uint128,uint256,uint256)",
))

var burnTopic = crypto.Keccak256Hash([]byte(
	"Burn(address,int24,int24,uint128,uint256,uint256)",
))

// LiquidityEvent - одно добавление (Mint) или изъятие (Burn) ликвидности
type LiquidityEvent struct {
	Block   uint64
	Added   bool     // true - Mint (добавили), false - Burn (забрали)
	Amount0 *big.Int // USDC, 6 decimals
	Amount1 *big.Int // WETH, 18 decimals
}

// FetchLiquidityEvents скачивает все Mint и Burn пула в блоках from..to (включительно) по порядку
func FetchLiquidityEvents(ctx context.Context, client *ethclient.Client, pool common.Address, from, to uint64) ([]LiquidityEvent, error) {
	var events []LiquidityEvent
	for start := from; start <= to; start += logChunkSize {
		end := min(start+logChunkSize-1, to)

		logs, err := client.FilterLogs(ctx, ethereum.FilterQuery{
			FromBlock: new(big.Int).SetUint64(start),
			ToBlock:   new(big.Int).SetUint64(end),
			Addresses: []common.Address{pool},
			// Mint ИЛИ Burn: оба вида событий одним запросом
			Topics: [][]common.Hash{{mintTopic, burnTopic}},
		})
		if err != nil {
			return nil, err
		}

		for _, lg := range logs {
			if lg.Topics[0] == mintTopic {
				// в Data лежат неиндексированные поля по порядку: sender, amount, amount0, amount1
				ev, err := liquidityABI.Unpack("Mint", lg.Data)
				if err != nil {
					return nil, err
				}
				events = append(events, LiquidityEvent{
					Block:   lg.BlockNumber,
					Added:   true,
					Amount0: ev[2].(*big.Int),
					Amount1: ev[3].(*big.Int),
				})
			} else {
				// в Data: amount, amount0, amount1
				ev, err := liquidityABI.Unpack("Burn", lg.Data)
				if err != nil {
					return nil, err
				}
				events = append(events, LiquidityEvent{
					Block:   lg.BlockNumber,
					Added:   false,
					Amount0: ev[1].(*big.Int),
					Amount1: ev[2].(*big.Int),
				})
			}
		}
	}
	return events, nil
}
