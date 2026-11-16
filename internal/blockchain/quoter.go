// Price impact: спрашивает официальный контракт Uniswap QuoterV2 что будет если прямо сейчас совершить обмен
// Quoter типо выполняет настоящий код обмена пула (со всеми тиками) и откатывает его
// в блокчейне ничего не меняется (eth_call)

package blockchain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// QuoterV2 в Ethereum mainnet, его factory() = 0x1F98431c8aD98523631AE4a59f267346ea31F984 (фабрика Uniswap V3)
var quoterV2 = common.HexToAddress("0x61fFE014bA17989E743c5F6cB21bF9697530B21e")

// quoteExactInputSingle - "сколько получу, если отдам ровно amountIn":
// вход (одна структура params):
// - tokenIn - какую монету отдаем пулу
// - tokenOut - какую монету получаем
// - amountIn - сколько отдаем, "сырое" число (USDC * 10^6, WETH * 10^18)
// - fee - ставка пула (500 = 0.05%), вместе с парой монет указывает на конкретный пул
// - sqrtPriceLimitX96 - до какой цены можно дойти, 0 = без ограничения
// выход:
// - amountOut - сколько получили бы уже за вычетом комиссии пула
// - sqrtPriceX96After - цена пула после обмена, в том же виде, что в slot0
// - initializedTicksCrossed - сколько границ тиков пересек обмен
// - gasEstimate - сколько газа стоил бы обмен (нам не нужен)
var quoterABI = mustABI(`[
{
	"name": "quoteExactInputSingle",
	"type": "function",
	"stateMutability": "nonpayable",
	"inputs": [{
		"name": "params",
		"type": "tuple",
		"components": [
			{"name": "tokenIn", "type": "address"},
			{"name": "tokenOut", "type": "address"},
			{"name": "amountIn", "type": "uint256"},
			{"name": "fee", "type": "uint24"},
			{"name": "sqrtPriceLimitX96", "type": "uint160"}
		]
	}],
	"outputs": [
		{"name": "amountOut", "type": "uint256"},
		{"name": "sqrtPriceX96After", "type": "uint160"},
		{"name": "initializedTicksCrossed", "type": "uint32"},
		{"name": "gasEstimate", "type": "uint256"}
	]
}
]`)

type Quote struct {
	AmountOut         *big.Int // сколько монет tokenOut получили бы, "сырое" число
	SqrtPriceX96After *big.Int // цена пула после обмена
	TicksCrossed      uint32   // сколько границ тиков пересек обмен
}

// QuoteExactInput: что будет, если отдать пулу amountIn монеты tokenIn и получить tokenOut
func QuoteExactInput(ctx context.Context, client *ethclient.Client, tokenIn, tokenOut common.Address, fee uint32, amountIn *big.Int) (Quote, error) {
	params := struct {
		TokenIn           common.Address
		TokenOut          common.Address
		AmountIn          *big.Int
		Fee               *big.Int
		SqrtPriceLimitX96 *big.Int
	}{tokenIn, tokenOut, amountIn, big.NewInt(int64(fee)), big.NewInt(0)}

	data, err := quoterABI.Pack("quoteExactInputSingle", params)
	if err != nil {
		return Quote{}, err
	}
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &quoterV2, Data: data}, nil)
	if err != nil {
		return Quote{}, err
	}
	values, err := quoterABI.Unpack("quoteExactInputSingle", out)
	if err != nil {
		return Quote{}, err
	}
	return Quote{
		AmountOut:         values[0].(*big.Int),
		SqrtPriceX96After: values[1].(*big.Int),
		TicksCrossed:      values[2].(uint32),
	}, nil
}
