package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"trading/internal/blockchain"
	"trading/internal/metrics"
	"trading/internal/pool"

	"github.com/ethereum/go-ethereum/common"
)

// Uniswap V3 USDC/WETH
const poolAddress = "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640"

// ~12 сек на блок -> ~300 блоков ~= 1 час
const blocksPerHour = 300

func main() {
	ctx := context.Background()

	client, err := blockchain.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	poolAddr := common.HexToAddress(poolAddress)

	// без текущей цены остальные метрики считаются, поэтому ошибку не роняем а печатаем в пункте 1
	sqrtNow, priceErr := blockchain.CurrentSqrtPriceX96(ctx, client, poolAddr)

	// для TVL: сколько монет лежит на адресе пула, ошибку тоже не роняем а печатаем в пункте 10
	// адреса монет спрашиваем у самого пула: в этом пуле token0 = USDC, token1 = WETH
	var usdcRaw, wethRaw *big.Int
	var usdcErr, wethErr error
	token0, token1, tokensErr := blockchain.PoolTokens(ctx, client, poolAddr)
	if tokensErr == nil {
		usdcRaw, usdcErr = blockchain.TokenBalance(ctx, client, token0, poolAddr)
		wethRaw, wethErr = blockchain.TokenBalance(ctx, client, token1, poolAddr)
	}

	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	toBlock := header.Number.Uint64()
	fromBlock := toBlock - blocksPerHour

	swaps, err := blockchain.FetchSwaps(ctx, client, poolAddr, fromBlock, toBlock)
	if err != nil {
		log.Fatal(err)
	}

	amounts := make([]*big.Int, 0, len(swaps))
	// цена после последнего обмена в каждом блоке: обмены внутри одного блока (например, сэндвичи ботов)
	// мгновенно дёргают цену туда-обратно и завышали бы волатильность
	var blockPrices []float64
	for i, s := range swaps {
		amounts = append(amounts, s.Amount0)

		price := pool.ETHPriceUSDC(s.SqrtPriceX96)
		if i > 0 && s.Block == swaps[i-1].Block {
			blockPrices[len(blockPrices)-1] = price
		} else {
			blockPrices = append(blockPrices, price)
		}
	}
	volume := metrics.VolumeUSDC(amounts)
	flow := metrics.BuySellUSDC(amounts)

	// выводим пул
	fmt.Printf("Пул USDC/WETH (Uniswap V3) %s\n", poolAddress)
	// вывыодим 1 пункт
	fmt.Println("Сейчас:")
	if priceErr != nil {
		fmt.Printf("1. Цена ETH: ошибка — %v\n", priceErr)
	} else {
		fmt.Printf("1. Цена ETH: %.2f USDC\n", pool.ETHPriceUSDC(sqrtNow))
	}
	// выводим 2 пункт
	fmt.Printf("За последний час (блоки %d .. %d):\n", fromBlock, toBlock)
	if len(swaps) > 0 {
		firstPrice := pool.ETHPriceUSDC(swaps[0].SqrtPriceX96)
		lastPrice := pool.ETHPriceUSDC(swaps[len(swaps)-1].SqrtPriceX96)
		fmt.Printf("2. Доходность цены: %+.4f%% (%.2f -> %.2f USDC, логарифмическая %+.6f)\n",
			metrics.SimpleReturn(firstPrice, lastPrice)*100,
			firstPrice, lastPrice,
			metrics.LogReturn(firstPrice, lastPrice))
	} else {
		fmt.Println("2. Доходность цены: нет обменов за час")
	}
	// выводим 3 пункт
	if len(blockPrices) > 1 {
		vol := metrics.RealizedVol(metrics.StepLogReturns(blockPrices))
		fmt.Printf("3. Реализованная волатильность: %.4f%% за час (в пересчёте на год %.1f%%, по %d блокам с обменами)\n",
			vol*100, metrics.AnnualizeHourlyVol(vol)*100, len(blockPrices))
	} else {
		fmt.Println("3. Реализованная волатильность: мало блоков с обменами")
	}
	// выводим 4 пункт
	fmt.Printf("4. Volume (Объем торгов): %.2f USDC\n", volume)
	// выводим 5 пункт
	fmt.Printf("5. Swap count: %d\n", flow.Count)
	// выводим 6 пункт
	fmt.Printf("6. Средний swap size: %.2f USDC\n", metrics.AvgSwapUSDC(volume, flow.Count))
	// выводим 7 пункт
	fmt.Printf("7. Buy volume: %.2f USDC\n", flow.BuyUSDC)
	// выводим 8 пункт
	fmt.Printf("8. Sell volume: %.2f USDC\n", flow.SellUSDC)
	// выводим 9 пункт
	fmt.Printf("9. Order-flow imbalance: %+.3f\n",
		metrics.OrderFlowImbalance(flow.BuyUSDC, flow.SellUSDC))
	// выводим 10 пункт
	switch {
	case priceErr != nil:
		fmt.Printf("10. TVL сейчас: нет текущей цены — %v\n", priceErr)
	case tokensErr != nil:
		fmt.Printf("10. TVL сейчас: не удалось узнать монеты пула — %v\n", tokensErr)
	case usdcErr != nil:
		fmt.Printf("10. TVL сейчас: ошибка баланса USDC — %v\n", usdcErr)
	case wethErr != nil:
		fmt.Printf("10. TVL сейчас: ошибка баланса WETH — %v\n", wethErr)
	default:
		tvl := metrics.PoolTVL(usdcRaw, wethRaw, pool.ETHPriceUSDC(sqrtNow))
		fmt.Printf("10. TVL сейчас: %.2f USDC (в пуле %.2f USDC и %.4f ETH ≈ %.2f USDC)\n",
			tvl.TotalUSD, tvl.USDC, tvl.WETH, tvl.WETHinUSD)
	}
}
