// Ставка комиссии пула: спрашивает у пула функцию fee()
// Ответ в миллионных долях от сделки: 500 = 500 / 1 000 000 = 0.05%
// И долю протокола Uniswap в этой комиссии: поле feeProtocol из slot0()

package blockchain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var poolFeeABI = mustABI(`[
{"name": "fee", "type": "function", "stateMutability": "view", "inputs": [], "outputs": [{"name": "", "type": "uint24"}]}
]`)

// PoolFee возвращает ставку комиссии пула в миллионных долях (500 = 0.05%)
func PoolFee(ctx context.Context, client *ethclient.Client, pool common.Address) (uint32, error) {
	data, err := poolFeeABI.Pack("fee")
	if err != nil {
		return 0, err
	}
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &pool, Data: data}, nil)
	if err != nil {
		return 0, err
	}
	values, err := poolFeeABI.Unpack("fee", out)
	if err != nil {
		return 0, err
	}
	return uint32(values[0].(*big.Int).Uint64()), nil
}

// PoolFeeProtocol возвращает поле feeProtocol из slot0: какую часть комиссии забирает протокол Uniswap
// младшие 4 бита - для комиссий в token0, старшие 4 бита - для комиссий в token1
// значение N: 0 = протокол ничего не берет, иначе протокол берет 1/N комиссии
func PoolFeeProtocol(ctx context.Context, client *ethclient.Client, pool common.Address) (uint8, error) {
	data, err := slot0ABI.Pack("slot0")
	if err != nil {
		return 0, err
	}
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &pool, Data: data}, nil)
	if err != nil {
		return 0, err
	}
	values, err := slot0ABI.Unpack("slot0", out)
	if err != nil {
		return 0, err
	}
	return values[5].(uint8), nil
}
