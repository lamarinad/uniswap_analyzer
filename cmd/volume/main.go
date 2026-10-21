package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"trading/internal/observe"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const poolAddres = "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640"

// ~12 сек на блок -> ~300 блоков ~= 1 час
const blocksPerHour = 300
const logChunkSize = 40

func main() {
	ctx := context.Background()

	client, err := ethclient.Dial("https://rpc.mevblocker.io")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	toBlock := header.Number.Uint64()
	fromBlock := toBlock - blocksPerHour

	swapABI, err := abi.JSON(strings.NewReader(`[
	{
		"type": "event", 
		"name": "Swap",
		"inputs": [
			{"name": "sender", "type": "address", "indexed": true},
			{"name": "recipient", "type": "address", "indexed": true},
			{"name": "amount0", "type": "int256", "indexed": false},
			{"name": "amount1", "type": "int256", "indexed": false},
			{"name": "sqrtPriceX96", "type": "uint160", "indexed": false},
			{"name": "liquidity", "type": "uint128", "indexed": false},
			{"name": "tick", "type": "int24", "indexed": false}
		]
	}
	]`))
	if err != nil {
		log.Fatal(err)
	}

	topic0 := crypto.Keccak256Hash([]byte(
		"Swap(address,address,int256,int256,uint160,uint128,int24)",
	))
	pool := common.HexToAddress(poolAddres)

	var amounts []*big.Int
	for start := fromBlock; start <= toBlock; start += logChunkSize {
		end := start + logChunkSize - 1
		if end > toBlock {
			end = toBlock
		}

		logs, err := client.FilterLogs(ctx, ethereum.FilterQuery{
			FromBlock: new(big.Int).SetUint64(start),
			ToBlock:   new(big.Int).SetUint64(end),
			Addresses: []common.Address{pool},
			Topics:    [][]common.Hash{{topic0}},
		})
		if err != nil {
			log.Fatal(err)
		}

		for _, lg := range logs {
			ev, err := swapABI.Unpack("Swap", lg.Data)
			if err != nil {
				log.Fatal(err)
			}
			amount0 := ev[0].(*big.Int)
			amounts = append(amounts, amount0)
		}
	}

	volume := observe.VolumeUSDC(amounts)
	fmt.Printf("swaps: %d\n", len(amounts))
	fmt.Printf("volume ~1h: %.2f USDC\n", volume)
	fmt.Printf("blocks: %d .. %d\n", fromBlock, toBlock)

	flow := observe.BuySellUSDC(amounts)
	avg := observe.AvgSwapUSDC(volume, flow.Count)
	fmt.Printf("buy ETH  ~1h: %.2f USDC\n", flow.BuyUSDC)
	fmt.Printf("sell ETH ~1h: %.2f USDC\n", flow.SellUSDC)
	fmt.Printf("avg swap: %.2f USDC\n", avg)
}
