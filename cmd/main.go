package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"trading/internal/blockchain"
	"trading/internal/exchanges"
	"trading/internal/metrics"
	"trading/internal/pool"

	"github.com/ethereum/go-ethereum/common"
)

// Uniswap V3 USDC/WETH
const poolAddress = "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640"

// ~12 сек на блок -> ~300 блоков ~= 1 час
const blocksPerHour = 300

// пункт 18: обмен от этого размера в USDC считаем крупным ("китом")
const whaleThresholdUSDC = 100_000

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

	// для пункта 15: лучшие заявки на биржах сразу после цены пула, чтобы сравнивать цены одного момента
	binance, binanceErr := exchanges.BinanceBidAsk(ctx, "ETHUSDC")
	bybit, bybitErr := exchanges.BybitBidAsk(ctx, "ETHUSDC")

	// для TVL: сколько монет лежит на адресе пула, ошибку тоже не роняем а печатаем в пункте 10
	// адреса монет спрашиваем у самого пула: в этом пуле token0 = USDC, token1 = WETH
	var usdcRaw, wethRaw *big.Int
	var usdcErr, wethErr error
	token0, token1, tokensErr := blockchain.PoolTokens(ctx, client, poolAddr)
	if tokensErr == nil {
		usdcRaw, usdcErr = blockchain.TokenBalance(ctx, client, token0, poolAddr)
		wethRaw, wethErr = blockchain.TokenBalance(ctx, client, token1, poolAddr)
	}

	// для пункта 12: ставка комиссии пула (500 = 0.05%) и доля протокола Uniswap, ошибки печатаем в пункте 12
	feePips, feeErr := blockchain.PoolFee(ctx, client, poolAddr)
	feeProtocol, feeProtocolErr := blockchain.PoolFeeProtocol(ctx, client, poolAddr)

	// для пункта 14: на каждый размер сделки спрашиваем Quoter, что будет при покупке ETH за USDC
	// и при продаже ETH на ту же сумму по текущей цене; ошибку печатаем в пункте 14
	impactSizesUSD := []float64{10_000, 100_000, 1_000_000}
	var impacts []impactQuote
	var impactErr error
	switch {
	case priceErr != nil:
		impactErr = fmt.Errorf("нет текущей цены: %w", priceErr)
	case tokensErr != nil:
		impactErr = fmt.Errorf("не удалось узнать монеты пула: %w", tokensErr)
	case feeErr != nil:
		impactErr = fmt.Errorf("не удалось узнать ставку комиссии: %w", feeErr)
	default:
		ethPrice := pool.ETHPriceUSDC(sqrtNow)
		for _, usd := range impactSizesUSD {
			usdcIn := toRaw(usd, 1e6)
			buy, err := blockchain.QuoteExactInput(ctx, client, token0, token1, feePips, usdcIn)
			if err != nil {
				impactErr = err
				break
			}
			wethIn := toRaw(usd/ethPrice, 1e18)
			sell, err := blockchain.QuoteExactInput(ctx, client, token1, token0, feePips, wethIn)
			if err != nil {
				impactErr = err
				break
			}
			impacts = append(impacts, impactQuote{
				usd:       usd,
				buy:       metrics.ImpactTrade{USDCRaw: usdcIn, WETHRaw: buy.AmountOut, PriceAfter: pool.ETHPriceUSDC(buy.SqrtPriceX96After)},
				sell:      metrics.ImpactTrade{USDCRaw: sell.AmountOut, WETHRaw: wethIn, PriceAfter: pool.ETHPriceUSDC(sell.SqrtPriceX96After)},
				buyTicks:  buy.TicksCrossed,
				sellTicks: sell.TicksCrossed,
			})
		}
	}

	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	toBlock := header.Number.Uint64()
	fromBlock := toBlock - blocksPerHour

	// для пункта 16: base fee следующего блока (туда попадет транзакция, отправленная сейчас)
	// и чаевые последних блоков
	gasFees, gasErr := blockchain.GasFees(ctx, client, header.Number)

	swaps, err := blockchain.FetchSwaps(ctx, client, poolAddr, fromBlock, toBlock)
	if err != nil {
		log.Fatal(err)
	}

	// для пункта 11: события Mint (добавили деньги) и Burn (забрали) за тот же час, ошибку печатаем в пункте 11
	liqEvents, liqErr := blockchain.FetchLiquidityEvents(ctx, client, poolAddr, fromBlock, toBlock)
	liqChanges := make([]metrics.LiquidityChange, 0, len(liqEvents))
	for _, e := range liqEvents {
		liqChanges = append(liqChanges, metrics.LiquidityChange{Added: e.Added, Amount0: e.Amount0, Amount1: e.Amount1})
	}

	amounts := make([]*big.Int, 0, len(swaps))
	// цена после последнего обмена в каждом блоке: обмены внутри одного блока (например, сэндвичи ботов)
	// мгновенно дергают цену туда-обратно и завышали бы волатильность
	var blockPrices []float64
	feeSwaps := make([]metrics.FeeSwap, 0, len(swaps))
	// для пункта 17: разные транзакции с обменами (в одной транзакции может быть несколько обменов)
	var txHashes []common.Hash
	seenTx := make(map[common.Hash]bool)
	for i, s := range swaps {
		amounts = append(amounts, s.Amount0)
		if !seenTx[s.TxHash] {
			seenTx[s.TxHash] = true
			txHashes = append(txHashes, s.TxHash)
		}

		price := pool.ETHPriceUSDC(s.SqrtPriceX96)
		feeSwaps = append(feeSwaps, metrics.FeeSwap{Amount0: s.Amount0, Amount1: s.Amount1, ETHPriceUSDC: price})
		if i > 0 && s.Block == swaps[i-1].Block {
			blockPrices[len(blockPrices)-1] = price
		} else {
			blockPrices = append(blockPrices, price)
		}
	}
	volume := metrics.VolumeUSDC(amounts)
	flow := metrics.BuySellUSDC(amounts)

	// для пункта 17: кто отправил каждую транзакцию (from), ошибку печатаем в пункте 17
	txSenders, sendersErr := blockchain.TxSenders(ctx, client, txHashes)

	// для пункта 19: время блоков с обменами и минутные свечи бирж за тот же час
	var swapBlocks []uint64
	seenBlock := make(map[uint64]bool)
	for _, s := range swaps {
		if !seenBlock[s.Block] {
			seenBlock[s.Block] = true
			swapBlocks = append(swapBlocks, s.Block)
		}
	}
	blockTime, blockTimeErr := blockchain.BlockTimestamps(ctx, client, swapBlocks)
	hourStart := time.Unix(int64(header.Time), 0).Add(-time.Duration(blocksPerHour)*12*time.Second - 2*time.Minute)
	hourEnd := time.Unix(int64(header.Time), 0).Add(time.Minute)
	binanceMin, binanceMinErr := exchanges.BinanceMinutePrices(ctx, "ETHUSDC", hourStart, hourEnd)
	bybitMin, bybitMinErr := exchanges.BybitMinutePrices(ctx, "ETHUSDC", hourStart, hourEnd)
	var arbSwaps []metrics.ArbSwap
	var arbErr error
	switch {
	case feeErr != nil:
		arbErr = fmt.Errorf("нет комиссии пула: %w", feeErr)
	case blockTimeErr != nil:
		arbErr = fmt.Errorf("нет времени блоков: %w", blockTimeErr)
	case binanceMinErr != nil && bybitMinErr != nil:
		arbErr = fmt.Errorf("Binance: %v; Bybit: %v", binanceMinErr, bybitMinErr)
	default:
		// цена до обмена - цена после предыдущего; у первого обмена часа ее нет
		var prev float64
		for i, s := range swaps {
			price := pool.ETHPriceUSDC(s.SqrtPriceX96)
			if i > 0 {
				ext, ok := exchanges.ExternalMinutePrice(binanceMin, bybitMin, blockTime[s.Block])
				if ok {
					arbSwaps = append(arbSwaps, metrics.ArbSwap{
						Amount0: s.Amount0, PriceBefore: prev, PriceAfter: price, ExternalPrice: ext,
					})
				}
			}
			prev = price
		}
	}

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
		fmt.Printf("3. Реализованная волатильность: %.4f%% за час (в пересчете на год %.1f%%, по %d блокам с обменами)\n",
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
	// tvlUSD запоминаем для пункта 13, при ошибке остается 0
	var tvlUSD float64
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
		tvlUSD = tvl.TotalUSD
		fmt.Printf("10. TVL сейчас: %.2f USDC (в пуле %.2f USDC и %.4f ETH ≈ %.2f USDC)\n",
			tvl.TotalUSD, tvl.USDC, tvl.WETH, tvl.WETHinUSD)
	}
	// выводим 11 пункт
	switch {
	case liqErr != nil:
		fmt.Printf("11. Net liquidity flow: ошибка загрузки Mint/Burn — %v\n", liqErr)
	case priceErr != nil:
		fmt.Printf("11. Net liquidity flow: нет текущей цены — %v\n", priceErr)
	default:
		lf := metrics.NetLiquidityFlow(liqChanges, pool.ETHPriceUSDC(sqrtNow))
		fmt.Printf("11. Net liquidity flow: %+.2f USDC (добавлений %d: %.2f USDC и %.4f ETH; изъятий %d: %.2f USDC и %.4f ETH)\n",
			lf.NetUSD, lf.Adds, lf.AddedUSDC, lf.AddedWETH, lf.Removes, lf.RemovedUSDC, lf.RemovedWETH)
	}
	// выводим 12 пункт
	switch {
	case feeErr != nil:
		fmt.Printf("12. Fees: не удалось узнать ставку комиссии пула — %v\n", feeErr)
	case feeProtocolErr != nil:
		fmt.Printf("12. Fees: не удалось узнать долю протокола — %v\n", feeProtocolErr)
	default:
		fees := metrics.PoolFees(feeSwaps, feePips, feeProtocol)
		fmt.Printf("12. Fees поставщикам: %.2f USDC (ставка %.2f%%, протокол берет: с USDC %s, с WETH %s; в USDC %.2f, в WETH %.6f ETH ≈ %.2f USDC)\n",
			fees.TotalUSD, float64(feePips)/1e4, protocolCut(feeProtocol%16), protocolCut(feeProtocol>>4), fees.USDC, fees.WETH, fees.WETHinUSD)
	}
	// выводим 13 пункт
	if ratio, ok := metrics.VolumeTVL(volume, tvlUSD); ok {
		fmt.Printf("13. Volume/TVL: %.4f (за час через обмены прошло %.2f%% денег пула)\n", ratio, ratio*100)
	} else {
		fmt.Println("13. Volume/TVL: нет TVL, причина в пункте 10")
	}
	// выводим 14 пункт
	if impactErr != nil {
		fmt.Printf("14. Price impact: ошибка — %v\n", impactErr)
	} else {
		priceNow := pool.ETHPriceUSDC(sqrtNow)
		fmt.Println("14. Price impact:")
		for _, q := range impacts {
			buy := metrics.PriceImpact(priceNow, q.buy)
			sell := metrics.PriceImpact(priceNow, q.sell)
			fmt.Printf("   купить ETH на %.0f USDC (%.4f ETH): цена %.2f -> %.2f (%+.4f%%), средняя цена %.2f, тиков пересечено %d\n",
				q.usd, buy.ETH, priceNow, q.buy.PriceAfter, buy.PriceChange*100, buy.AvgPrice, q.buyTicks)
			fmt.Printf("   продать ETH на %.0f USDC (%.4f ETH): цена %.2f -> %.2f (%+.4f%%), средняя цена %.2f, тиков пересечено %d\n",
				q.usd, sell.ETH, priceNow, q.sell.PriceAfter, sell.PriceChange*100, sell.AvgPrice, q.sellTicks)
		}
	}
	// выводим 15 пункт
	if priceErr != nil {
		fmt.Printf("15. Отклонение от внешней цены: нет цены пула — %v\n", priceErr)
	} else {
		poolPrice := pool.ETHPriceUSDC(sqrtNow)
		fmt.Printf("15. Отклонение от внешней цены (пул %.2f USDC:\n", poolPrice)
		for _, ex := range []struct {
			name string
			q    exchanges.BidAsk
			err  error
		}{
			{"Binance ETH/USDC", binance, binanceErr},
			{"Bybit ETH/USDC", bybit, bybitErr},
		} {
			if ex.err != nil {
				fmt.Printf("   %s: ошибка — %v\n", ex.name, ex.err)
				continue
			}
			mid := metrics.MidPrice(ex.q.Bid, ex.q.Ask)
			fmt.Printf("   %s %.2f (bid %.2f / ask %.2f): %+.4f%%\n",
				ex.name, mid, ex.q.Bid, ex.q.Ask, metrics.ExternalPriceDeviation(poolPrice, mid)*100)
		}
	}
	// выводим 16 пункт
	switch {
	case gasErr != nil:
		// без eth_feeHistory показываем хотя бы base fee последнего блока из уже загруженного заголовка
		fmt.Printf("16. Gas price: base fee последнего блока %.4f gwei (base fee следующего блока и чаевые узнать не удалось — %v)\n",
			metrics.GasPrice(header.BaseFee, nil, 0).BaseGwei, gasErr)
	case priceErr != nil:
		gas := metrics.GasPrice(gasFees.NextBaseFee, gasFees.Tips, 0)
		fmt.Printf("16. Gas price: %.4f gwei (base fee %.4f + чаевые %.4f; обмен на Uniswap ~%d газа ≈ %.6f ETH, в доллары не перевести — нет цены)\n",
			gas.TotalGwei, gas.BaseGwei, gas.TipGwei, metrics.SwapGas, gas.SwapCostETH)
	default:
		gas := metrics.GasPrice(gasFees.NextBaseFee, gasFees.Tips, pool.ETHPriceUSDC(sqrtNow))
		fmt.Printf("16. Gas price: %.4f gwei (base fee %.4f + чаевые %.4f; обмен на Uniswap ~%d газа ≈ %.6f ETH ≈ %.2f USDC)\n",
			gas.TotalGwei, gas.BaseGwei, gas.TipGwei, metrics.SwapGas, gas.SwapCostETH, gas.SwapCostUSD)
	}
	// выводим 17 пункт
	if sendersErr != nil {
		fmt.Printf("17. Уникальные traders: не удалось узнать отправителей транзакций — %v\n", sendersErr)
	} else {
		fmt.Printf("17. Уникальные traders: %d (разных адресов-отправителей; транзакций с обменами: %d)\n",
			metrics.UniqueTraders(txSenders), len(txHashes))
	}
	// выводим 18 пункт
	whale := metrics.WhaleVolume(amounts, whaleThresholdUSDC)
	fmt.Printf("18. Whale volume: %.2f USDC (сделок от %.0f USDC: %d; это %.1f%% всего объема)\n",
		whale.VolumeUSDC, float64(whaleThresholdUSDC), whale.Count, whale.Share*100)
	// выводим 19 пункт
	if arbErr != nil {
		fmt.Printf("19. Arbitrage volume: %v\n", arbErr)
	} else {
		arb := metrics.ArbitrageVolume(arbSwaps, float64(feePips)/1e6, volume)
		src := "Binance и Bybit"
		switch {
		case binanceMinErr != nil:
			src = fmt.Sprintf("Bybit (Binance: %v)", binanceMinErr)
		case bybitMinErr != nil:
			src = fmt.Sprintf("Binance (Bybit: %v)", bybitMinErr)
		}
		fmt.Printf("19. Arbitrage volume: %.2f USDC (%d сделок; %.1f%% всего объема; к цене %s, расхождение до сделки больше комиссии пула)\n",
			arb.VolumeUSDC, arb.Count, arb.Share*100, src)
	}
}

// impactQuote - ответы Quoter для одного размера сделки (пункт 14)
type impactQuote struct {
	usd                 float64
	buy, sell           metrics.ImpactTrade
	buyTicks, sellTicks uint32
}

// toRaw - перевод суммы в сырое число монеты: 1.5 ETH -> 1.5 * 10^18
func toRaw(amount, decimals float64) *big.Int {
	raw, _ := new(big.Float).Mul(big.NewFloat(amount), big.NewFloat(decimals)).Int(nil)
	return raw
}

// protocolCut - подпись доли протокола для вывода: 0 = "ничего", N = "1/N"
func protocolCut(n uint8) string {
	if n == 0 {
		return "ничего"
	}
	return fmt.Sprintf("1/%d", n)
}
