package main

import (
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	RPCURL        string   `yaml:"rpc_url"`
	Wallets       []string `yaml:"wallets"`
	PollInterval  int      `yaml:"poll_interval"`
	AIAnalyzerURL string   `yaml:"ai_analyzer_url,omitempty"`
	DatabaseURL   string   `yaml:"database_url,omitempty"`
}

func loadConfig() (*Config, error) {
	// First try environment variables
	rpcURL := os.Getenv("RPC_URL")
	aiAnalyzerURL := os.Getenv("AI_ANALYZER_URL")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("POSTGRES_DSN")
	}

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
			DatabaseURL:   dbURL,
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
