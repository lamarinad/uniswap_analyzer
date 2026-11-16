// Метрика 17 "Уникальные traders"
// Сколько разных адресов совершали обмены за период
// трейдер - кошелек, отправивший транзакцию (from); у посредников (UniswapX, CoW Swap) это адрес посредника, а не человека

package metrics

import "github.com/ethereum/go-ethereum/common"

// UniqueTraders: senders - from каждой транзакции с обменами в пуле (по одному на транзакцию)
func UniqueTraders(senders []common.Address) int {
	seen := make(map[common.Address]bool, len(senders))
	for _, s := range senders {
		seen[s] = true
	}
	return len(seen)
}
