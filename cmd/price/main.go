package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"trading/internal/pool"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Uniswap V3
const poolAddres = "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640"

func main() {
	ctx := context.Background()

	client, err := ethclient.Dial("https://ethereum.publicnode.com")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	poolABI, err := abi.JSON(strings.NewReader(`[
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
	]`))
	if err != nil {
		log.Fatal(err)
	}

	data, err := poolABI.Pack("slot0")
	if err != nil {
		log.Fatal(err)
	}

	addr := common.HexToAddress(poolAddres)
	out, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &addr,
		Data: data,
	}, nil)
	if err != nil {
		log.Fatal(err)
	}

	values, err := poolABI.Unpack("slot0", out)
	if err != nil {
		log.Fatal(err)
	}

	sqrtPriceX96 := values[0].(*big.Int)
	price := pool.ETHPriceUSDC(sqrtPriceX96)

	fmt.Printf("ETH price ~= %.2f USDC\n", price)
}
