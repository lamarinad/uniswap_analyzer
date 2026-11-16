// Подключение к узлу Ethereum (RPC) - "справочной службе" блокчейна
// через которую мы читаем состояние пула и журнал его событий
// Здесь же общий помощник для разбора описаний контрактов (ABI)

package blockchain

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/ethclient"
)

// pc.mevblocker.io — адрес публичного бесплатного узла Ethereum mainnet
// от сервиса MEV Blocker
// через него сервис задает все вопросы блокчейну
const rpcURL = "https://rpc.mevblocker.io"

// Connect открывает соединение с узлом Ethereum
func Connect() (*ethclient.Client, error) {
	return ethclient.Dial(rpcURL)
}

func mustABI(s string) abi.ABI {
	parsed, err := abi.JSON(strings.NewReader(s))
	if err != nil {
		panic(err)
	}
	return parsed
}
