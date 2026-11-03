// Балансы монет: сколько монет (USDC, WETH и т.д.) лежит на адресе прямо сейчас
// Каждая монета - отдельный контракт, у него спрашиваем balanceOf(адрес)
// Возвращает "сырое" число без запятой: для USDC делить на 10^6, для WETH на 10^18

package blockchain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var erc20ABI = mustABI(`[
{
	"name": "balanceOf",
	"type": "function",
	"stateMutability": "view",
	"inputs": [{"name": "account", "type": "address"}],
	"outputs": [{"name": "", "type": "uint256"}]
}
]`)

// TokenBalance возвращает баланс монеты token на адресе holder
func TokenBalance(ctx context.Context, client *ethclient.Client, token, holder common.Address) (*big.Int, error) {
	data, err := erc20ABI.Pack("balanceOf", holder)
	if err != nil {
		return nil, err
	}
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &token, Data: data}, nil)
	if err != nil {
		return nil, err
	}
	values, err := erc20ABI.Unpack("balanceOf", out)
	if err != nil {
		return nil, err
	}
	return values[0].(*big.Int), nil
}
