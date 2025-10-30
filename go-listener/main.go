package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		log.Fatalf("Failed to connect to RPC: %v", err)
	}
	defer client.Close()

	fmt.Println("✅ Connected to Ethereum RPC node via Alchemy!")
	fmt.Println("👛 Monitoring wallets:", cfg.Wallets)
	if cfg.AIAnalyzerURL != "" {
		fmt.Println("🤖 AI Analyzer URL:", cfg.AIAnalyzerURL)
	} else {
		fmt.Println("⚠️  AI Analyzer URL not configured - transactions will only be logged")
	}

	// Load last processed block from state
	lastBlock, err := loadState("state.json")
	if err != nil {
		log.Printf("Error loading state, starting from block 0: %v", err)
		lastBlock = 0
	}

	fmt.Printf("Starting from block %d\n", lastBlock)

	// Main monitoring loop
	for {
		newLastBlock, err := fetchNewTransactions(client, cfg.Wallets, lastBlock, cfg.AIAnalyzerURL)
		if err != nil {
			log.Printf("Error fetching transactions: %v", err)
		} else if newLastBlock > lastBlock {
			// Save state if we processed new blocks
			err = saveState("state.json", newLastBlock)
			if err != nil {
				log.Printf("Error saving state: %v", err)
			}
			lastBlock = newLastBlock
			fmt.Printf("✅ Updated last processed block to %d\n", lastBlock)
		} else {
			fmt.Println("⏳ No new blocks to process")
		}

		fmt.Printf("💤 Sleeping for %d seconds...\n", cfg.PollInterval)
		time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
	}
}
