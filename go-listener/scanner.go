package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func fetchNewTransactions(client *ethclient.Client, wallets []string, lastBlock uint64, analyzerURL string) (uint64, error) {
	ctx := context.Background()

	latestHeader, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return lastBlock, err
	}
	latestBlock := latestHeader.Number.Uint64()

	if lastBlock == 0 && latestBlock > 1000 {
		lastBlock = latestBlock - 1000
		fmt.Printf("Starting from recent block: %d (latest: %d)\n", lastBlock, latestBlock)
	}

	if lastBlock >= latestBlock {
		return lastBlock, nil
	}

	walletSet := make(map[common.Address]bool)
	for _, w := range wallets {
		walletSet[common.HexToAddress(w)] = true
	}

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return lastBlock, err
	}
	signer := types.LatestSignerForChainID(chainID)

	for blockNum := lastBlock + 1; blockNum <= latestBlock; blockNum++ {
		block, err := client.BlockByNumber(ctx, new(big.Int).SetUint64(blockNum))
		if err != nil {
			log.Printf("Error fetching block %d: %v", blockNum, err)
			return lastBlock, err
		}

		fmt.Printf("Scanning block %d (%d transactions)\n", blockNum, len(block.Transactions()))

		foundCount := 0
		for _, tx := range block.Transactions() {
			from, err := types.Sender(signer, tx)
			if err != nil {
				continue
			}

			to := common.Address{}
			if tx.To() != nil {
				to = *tx.To()
			}

			if walletSet[from] || walletSet[to] {
				foundCount++
				txData := map[string]interface{}{
					"hash":  tx.Hash().Hex(),
					"from":  from.Hex(),
					"to":    to.Hex(),
					"value": tx.Value().String(),
					"gas":   tx.Gas(),
					"gasPrice": func() string {
						if tx.GasPrice() != nil {
							return tx.GasPrice().String()
						}
						return "0"
					}(),
					"blockNum":  blockNum,
					"timestamp": block.Time(),
					"input":     common.Bytes2Hex(tx.Data()),
				}

				jsonData, _ := json.Marshal(txData)
				fmt.Printf("Found relevant transaction: %s\n", string(jsonData))

				if analyzerURL != "" {
					if err := sendToAIAnalyzer(analyzerURL, txData); err != nil {
						log.Printf("Error sending to AI analyzer: %v", err)
					}
				}
			}
		}

		if foundCount > 0 {
			fmt.Printf("Found %d relevant transactions in block %d\n", foundCount, blockNum)
		}

		lastBlock = blockNum
	}

	return lastBlock, nil
}
