package metrics

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestUniqueTraders(t *testing.T) {
	a := common.HexToAddress("0xa")
	b := common.HexToAddress("0xb")
	c := common.HexToAddress("0xc")

	// 5 транзакций от A, B, A, C, A -> 3 разных трейдера
	if got := UniqueTraders([]common.Address{a, b, a, c, a}); got != 3 {
		t.Fatalf("got %d, want 3", got)
	}
	if got := UniqueTraders(nil); got != 0 {
		t.Fatalf("no swaps: got %d, want 0", got)
	}
}
