package metrics

import (
	"math"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestMEVActivity(t *testing.T) {
	bot := common.HexToAddress("0xb07")
	victim := common.HexToAddress("0x111")
	usdc := func(n int64) *big.Int { return big.NewInt(n * 1e6) }
	swap := func(block uint64, idx uint, from common.Address, amount int64) MEVSwap {
		return MEVSwap{Block: block, TxIndex: idx, From: from, Amount0: usdc(amount)}
	}

	got := MEVActivity([]MEVSwap{
		// бот купил, жертва купила следом, бот продал
		swap(1, 1, bot, 1_000),
		swap(1, 2, victim, 5_000),
		swap(1, 3, bot, -1_000),
		// тот же бот купил и сразу продал, чужой сделки между ними нет
		swap(2, 1, bot, 100),
		swap(2, 2, bot, -100),
	})
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
	if got.Sandwiches != 1 || !near(got.BotUSDC, 2_000) || !near(got.VictimUSDC, 5_000) {
		t.Fatalf("got %+v, want sandwiches=1 bot=2000 victim=5000", got)
	}
}
