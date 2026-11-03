// Текущая цена пула: читает "табло" пула (функцию slot0) прямо сейчас
// Возвращает цену в том виде, в каком ее хранит Uniswap (sqrtPriceX96)
// перевести ее в USDC за 1 ETH можно функцией pool.ETHPriceUSDC

package blockchain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var slot0ABI = mustABI(`[
{
	"name": "slot0",
	"type": "function",
	"stateMutability": "view",
	"inputs": [],
	"outputs": [
		{"name": "sqrtPriceX96", "type": "uint160"},
		{"name": "tick", "type": "int24"},
		{"name": "observationIndex", "type": "uint16"},
		{"name": "observationCardinality", "type": "uint16"},
		{"name": "observationCardinalityNext", "type": "uint16"},
		{"name": "feeProtocol", "type": "uint8"},
		{"name": "unlocked", "type": "bool"}
	]
}
]`)

// CurrentSqrtPriceX96 возвращает текущую закодированную цену пула
func CurrentSqrtPriceX96(ctx context.Context, client *ethclient.Client, pool common.Address) (*big.Int, error) {
	data, err := slot0ABI.Pack("slot0")
	if err != nil {
		return nil, err
	}
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &pool, Data: data}, nil)
	if err != nil {
		return nil, err
	}
	values, err := slot0ABI.Unpack("slot0", out)
	if err != nil {
		return nil, err
	}
	return values[0].(*big.Int), nil
}
