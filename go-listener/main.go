package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/yaml.v2"
)

type Config struct {
	RPCURL       string   `yaml:"rpc_url"`
	Wallets      []string `yaml:"wallets"`
	PollInterval int      `yaml:"poll_interval"`
	// We'll keep your existing structure and add AI analyzer as optional
	AIAnalyzerURL string `yaml:"ai_analyzer_url,omitempty"` // Optional field
}

type State struct {
	LastBlock uint64 `json:"last_block"`
}

// Add this function to load config from env or file
func loadConfig() (*Config, error) {
	// First try environment variables
	rpcURL := os.Getenv("RPC_URL")
	aiAnalyzerURL := os.Getenv("AI_ANALYZER_URL")

	if rpcURL != "" {
		// Use environment variables
		wallets := strings.Split(os.Getenv("WALLETS"), ",")
		if len(wallets) == 0 {
			wallets = []string{"0x1234567890abcdef1234567890abcdef12345678"}
		}

		pollInterval := 15
		if pi := os.Getenv("POLL_INTERVAL"); pi != "" {
			if piVal, err := strconv.Atoi(pi); err == nil {
				pollInterval = piVal
			}
		}

		return &Config{
			RPCURL:        rpcURL,
			Wallets:       wallets,
			PollInterval:  pollInterval,
			AIAnalyzerURL: aiAnalyzerURL,
		}, nil
	}

	// Fall back to config file
	return loadConfigFromFile("config.yaml")
}

func loadConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}

func loadState(path string) (uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // Start from block 0 if no state file
		}
		return 0, err
	}
	var state State
	json.Unmarshal(data, &state)
	return state.LastBlock, nil
}

func saveState(path string, blockNum uint64) error {
	state := State{LastBlock: blockNum}
	data, _ := json.Marshal(state)
	return os.WriteFile(path, data, 0644)
}

func sendToAIAnalyzer(analyzerURL string, txData map[string]interface{}) error {
	jsonData, err := json.Marshal(txData)
	if err != nil {
		return err
	}

	resp, err := http.Post(analyzerURL+"/analyze", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AI analyzer error: %s", string(body))
	}

	// Read and log the analysis result
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	log.Printf("Risk Analysis: %+v", result)

	return nil
}

func fetchNewTransactions(client *ethclient.Client, wallets []string, lastBlock uint64, analyzerURL string) (uint64, error) {
	ctx := context.Background()

	// Get latest block number
	latestHeader, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return lastBlock, err
	}
	latestBlock := latestHeader.Number.Uint64()

	// If no previous state, start from recent blocks (last 1000 blocks)
	if lastBlock == 0 && latestBlock > 1000 {
		lastBlock = latestBlock - 1000
		fmt.Printf("Starting from recent block: %d (latest: %d)\n", lastBlock, latestBlock)
	}

	// If we're already at the latest block, nothing to do
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

	// Scan blocks from lastBlock + 1 to latest
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

			// Check if transaction involves any monitored wallet
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

				// Send to AI analyzer if URL is configured
				if analyzerURL != "" {
					err := sendToAIAnalyzer(analyzerURL, txData)
					if err != nil {
						log.Printf("Error sending to AI analyzer: %v", err)
					}
				}
			}
		}

		if foundCount > 0 {
			fmt.Printf("Found %d relevant transactions in block %d\n", foundCount, blockNum)
		}

		lastBlock = blockNum // Update last processed block
	}

	return lastBlock, nil
}

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
