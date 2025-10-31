package main

import (
	"fmt"
	"log"
	"time"

	"context"
	"net/http"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v5/pgxpool"
	dbpkg "github.com/nidhish1/BlockSentinel/go-listener/db"
	routes "github.com/nidhish1/BlockSentinel/go-listener/routes"
	utilpkg "github.com/nidhish1/BlockSentinel/go-listener/util"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Optional: connect to Postgres if configured (with retry/backoff)
	var dbpool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		pool, dbErr := utilpkg.ConnectPostgresWithBackoff(context.Background(), cfg.DatabaseURL, 60*time.Second)
		if dbErr != nil {
			log.Printf("⚠️  Postgres unavailable: %v", dbErr)
		} else {
			log.Printf("✅ Connected to Postgres")
			// Run DB migrations at startup
			if err := utilpkg.RunMigrations(cfg.DatabaseURL, "./migrations"); err != nil {
				log.Printf("⚠️  Migrations failed: %v", err)
			} else {
				log.Printf("✅ Database migrations applied")
			}
			mux := http.NewServeMux()
			routes.RegisterRoutes(mux, pool)
			go func() {
				log.Printf("🌐 HTTP server listening on :8080")
				if err := http.ListenAndServe(":8080", mux); err != nil {
					log.Printf("HTTP server error: %v", err)
				}
			}()
			dbpool = pool
			defer pool.Close()
		}
	} else {
		log.Printf("ℹ️  DATABASE_URL not set; skipping Postgres connection")
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
		// Determine wallets source: prefer DB, fallback to config
		wallets := cfg.Wallets
		if dbpool != nil {
			if w, derr := dbpkg.FetchMonitoredWallets(context.Background(), dbpool); derr == nil && len(w) > 0 {
				wallets = w
			}
		}

		newLastBlock, err := fetchNewTransactions(client, wallets, lastBlock, cfg.AIAnalyzerURL)
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
