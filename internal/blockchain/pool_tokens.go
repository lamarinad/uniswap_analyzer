// Монеты пула: спрашивает у пула, из каких двух монет он состоит
// Функции пула token0() и token1() возвращают адреса контрактов этих монет

package blockchain

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var poolTokensABI = mustABI(`[
{"name": "token0", "type": "function", "stateMutability": "view", "inputs": [], "outputs": [{"name": "", "type": "address"}]},
{"name": "token1", "type": "function", "stateMutability": "view", "inputs": [], "outputs": [{"name": "", "type": "address"}]}
]`)

// PoolTokens возвращает адреса монет пула: token0 и token1
func PoolTokens(ctx context.Context, client *ethclient.Client, pool common.Address) (common.Address, common.Address, error) {
	token0, err := callAddress(ctx, client, pool, "token0")
	if err != nil {
		return common.Address{}, common.Address{}, err
	}
	token1, err := callAddress(ctx, client, pool, "token1")
	if err != nil {
		return common.Address{}, common.Address{}, err
	}
	return token0, token1, nil
}

func callAddress(ctx context.Context, client *ethclient.Client, pool common.Address, method string) (common.Address, error) {
	data, err := poolTokensABI.Pack(method)
	if err != nil {
		return common.Address{}, err
	}
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &pool, Data: data}, nil)
	if err != nil {
		return common.Address{}, err
	}
	values, err := poolTokensABI.Unpack(method, out)
	if err != nil {
		return common.Address{}, err
	}
	return values[0].(common.Address), nil
}
