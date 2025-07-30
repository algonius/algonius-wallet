// SPDX-License-Identifier: Apache-2.0
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
	"go.uber.org/zap"
)

// Config represents the main configuration structure
type Config struct {
	Wallet   WalletConfig   `yaml:"wallet"`
	Chains   ChainsConfig   `yaml:"chains"`
	DEX      DEXConfig      `yaml:"dex"`
	Security SecurityConfig `yaml:"security"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// WalletConfig contains wallet-specific settings
type WalletConfig struct {
	DataDir     string `yaml:"data_dir"`
	PrivateKey  string `yaml:"private_key,omitempty"`  // Base58 encoded private key
	NetworkMode string `yaml:"network_mode"`           // mainnet, testnet, devnet
}

// ChainsConfig contains blockchain network configurations
type ChainsConfig struct {
	Solana   SolanaChainConfig   `yaml:"solana"`
	Ethereum EthereumChainConfig `yaml:"ethereum"`
	BSC      BSCChainConfig      `yaml:"bsc"`
}

// SolanaChainConfig contains Solana-specific configuration
type SolanaChainConfig struct {
	Enabled       bool                    `yaml:"enabled"`
	RPCEndpoints  []string                `yaml:"rpc_endpoints"`
	WSEndpoint    string                  `yaml:"ws_endpoint"`
	Commitment    string                  `yaml:"commitment"`
	ReserveSOL    float64                 `yaml:"reserve_sol"`
	Retry         RetryConfig             `yaml:"retry"`
	Confirmation  ConfirmationConfig      `yaml:"confirmation"`
	Jito          JitoConfig              `yaml:"jito"`
	Broadcast     BroadcastConfig         `yaml:"broadcast"`
}

// EthereumChainConfig contains Ethereum-specific configuration
type EthereumChainConfig struct {
	Enabled      bool     `yaml:"enabled"`
	RPCEndpoints []string `yaml:"rpc_endpoints"`
	ChainID      int      `yaml:"chain_id"`
	GasStrategy  string   `yaml:"gas_strategy"`
}

// BSCChainConfig contains BSC-specific configuration
type BSCChainConfig struct {
	Enabled      bool     `yaml:"enabled"`
	RPCEndpoints []string `yaml:"rpc_endpoints"`
	ChainID      int      `yaml:"chain_id"`
	GasStrategy  string   `yaml:"gas_strategy"`
}

// RetryConfig defines retry behavior for failed transactions
type RetryConfig struct {
	MaxRetries           int           `yaml:"max_retries"`
	SlippageIncrementBps int           `yaml:"slippage_increment_bps"`
	BaseRetryDelay       time.Duration `yaml:"base_retry_delay"`
	MaxTotalSlippageBps  int           `yaml:"max_total_slippage_bps"`
	GasStrategy          string        `yaml:"gas_strategy"`
}

// ConfirmationConfig defines transaction confirmation behavior
type ConfirmationConfig struct {
	Timeout               time.Duration `yaml:"timeout"`
	PollInterval          time.Duration `yaml:"poll_interval"`
	RequiredConfirmations int           `yaml:"required_confirmations"`
}

// JitoConfig defines MEV protection settings
type JitoConfig struct {
	Enabled         bool   `yaml:"enabled"`
	BaseTipLamports uint64 `yaml:"base_tip_lamports"`
	MaxTipLamports  uint64 `yaml:"max_tip_lamports"`
	TipStrategy     string `yaml:"tip_strategy"`
	BundleEndpoint  string `yaml:"bundle_endpoint"`
}

// BroadcastConfig defines transaction broadcast settings
type BroadcastConfig struct {
	Channel  string            `yaml:"channel"`  // solana-rpc, okex, jito, jito-bundle, paper
	Channels []BroadcastChannel `yaml:"channels"`
}

// BroadcastChannel defines individual broadcast channel configuration
type BroadcastChannel struct {
	Name     string                 `yaml:"name"`
	Enabled  bool                   `yaml:"enabled"`
	Priority int                    `yaml:"priority"`
	Config   map[string]interface{} `yaml:"config"`
}

// DEXConfig contains DEX aggregator configurations
type DEXConfig struct {
	OKEx      OKExConfig      `yaml:"okex"`
	Jupiter   JupiterConfig   `yaml:"jupiter"`
	PumpFun   PumpFunConfig   `yaml:"pumpfun"`
	Composite CompositeConfig `yaml:"composite"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	RPM   int `yaml:"rpm"`   // Requests per minute
	Burst int `yaml:"burst"` // Burst capacity
}

// OKExConfig contains OKEx DEX configuration
type OKExConfig struct {
	Enabled          bool             `yaml:"enabled"`
	BroadcastChannel string           `yaml:"broadcast_channel"`
	APIKey           string           `yaml:"api_key,omitempty"`
	SecretKey        string           `yaml:"secret_key,omitempty"`
	PassPhrase       string           `yaml:"passphrase,omitempty"`
	ProjectID        string           `yaml:"project_id,omitempty"`
	BaseURL          string           `yaml:"base_url"`
	Timeout          int              `yaml:"timeout"`
	RateLimit        *RateLimitConfig `yaml:"rate_limit,omitempty"`
}

// JupiterConfig contains Jupiter aggregator configuration
type JupiterConfig struct {
	Enabled          bool   `yaml:"enabled"`
	BroadcastChannel string `yaml:"broadcast_channel"`
	BaseURL          string `yaml:"base_url"`
	Timeout          int    `yaml:"timeout"`
}

// PumpFunConfig contains PumpFun DEX configuration
type PumpFunConfig struct {
	Enabled          bool   `yaml:"enabled"`
	BroadcastChannel string `yaml:"broadcast_channel"`
	BaseURL          string `yaml:"base_url"`
	Timeout          int    `yaml:"timeout"`
}

// CompositeConfig contains composite DEX configuration
type CompositeConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Providers []string `yaml:"providers"`
	Strategy  string   `yaml:"strategy"` // best_price, fastest, balanced
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	EncryptPrivateKeys bool   `yaml:"encrypt_private_keys"`
	KeyDerivationPath  string `yaml:"key_derivation_path"`
	SessionTimeout     int    `yaml:"session_timeout"`
	RequirePassword    bool   `yaml:"require_password"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	OutputFile string `yaml:"output_file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Wallet: WalletConfig{
			DataDir:     "~/.algonius-wallet",
			NetworkMode: "mainnet",
		},
		Chains: ChainsConfig{
			Solana: SolanaChainConfig{
				Enabled: true,
				RPCEndpoints: []string{
					"https://api.mainnet-beta.solana.com",
					"https://solana-api.projectserum.com",
				},
				WSEndpoint: "wss://api.mainnet-beta.solana.com",
				Commitment: "confirmed",
				ReserveSOL: 0.01,
				Retry: RetryConfig{
					MaxRetries:           3,
					SlippageIncrementBps: 50,
					BaseRetryDelay:       2 * time.Second,
					MaxTotalSlippageBps:  1000,
					GasStrategy:          "level_up",
				},
				Confirmation: ConfirmationConfig{
					Timeout:               2 * time.Minute,
					PollInterval:          3 * time.Second,
					RequiredConfirmations: 1,
				},
				Jito: JitoConfig{
					Enabled:         false,
					BaseTipLamports: 1000,
					MaxTipLamports:  100000,
					TipStrategy:     "exponential",
					BundleEndpoint:  "https://mainnet.block-engine.jito.wtf/api/v1/bundles",
				},
				Broadcast: BroadcastConfig{
					Channel: "solana-rpc",
					Channels: []BroadcastChannel{
						{Name: "solana-rpc", Enabled: true, Priority: 1},
						{Name: "okex", Enabled: false, Priority: 2},
						{Name: "jito", Enabled: false, Priority: 3},
					},
				},
			},
			Ethereum: EthereumChainConfig{
				Enabled:      true,
				RPCEndpoints: []string{"https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY"},
				ChainID:      1,
				GasStrategy:  "fast",
			},
			BSC: BSCChainConfig{
				Enabled:      true,
				RPCEndpoints: []string{"https://bsc-dataseed.binance.org"},
				ChainID:      56,
				GasStrategy:  "standard",
			},
		},
		DEX: DEXConfig{
			OKEx: OKExConfig{
				Enabled:          false,
				BroadcastChannel: "okex",
				BaseURL:          "https://www.okx.com",
				Timeout:          30,
			},
			Jupiter: JupiterConfig{
				Enabled:          true,
				BroadcastChannel: "solana-rpc",
				BaseURL:          "https://quote-api.jup.ag/v6",
				Timeout:          10,
			},
			PumpFun: PumpFunConfig{
				Enabled:          false,
				BroadcastChannel: "solana-rpc",
				BaseURL:          "https://pumpportal.fun",
				Timeout:          10,
			},
			Composite: CompositeConfig{
				Enabled:   true,
				Providers: []string{"jupiter", "okex"},
				Strategy:  "best_price",
			},
		},
		Security: SecurityConfig{
			EncryptPrivateKeys: true,
			KeyDerivationPath:  "m/44'/501'/0'/0'",
			SessionTimeout:     3600,
			RequirePassword:    false,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			OutputFile: "~/.algonius-wallet/logs/wallet.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
		},
	}
}

// TestConfig returns a configuration suitable for testing
func TestConfig() *Config {
	config := DefaultConfig()
	
	// Use devnet for testing
	config.Wallet.NetworkMode = "devnet"
	config.Chains.Solana.RPCEndpoints = []string{"https://api.devnet.solana.com"}
	config.Chains.Solana.WSEndpoint = "wss://api.devnet.solana.com"
	config.Chains.Ethereum.RPCEndpoints = []string{"https://eth-goerli.g.alchemy.com/v2/test"}
	config.Chains.BSC.RPCEndpoints = []string{"https://data-seed-prebsc-1-s1.binance.org:8545"}
	
	// Faster confirmations and less aggressive retry for testing
	config.Chains.Solana.Confirmation.Timeout = 30 * time.Second
	config.Chains.Solana.Confirmation.PollInterval = 1 * time.Second
	config.Chains.Solana.Retry.MaxRetries = 1
	config.Chains.Solana.Retry.BaseRetryDelay = 500 * time.Millisecond
	
	// Use paper trading for DEX operations in tests
	config.DEX.OKEx.BroadcastChannel = "paper"
	config.DEX.Jupiter.BroadcastChannel = "paper"
	config.DEX.PumpFun.BroadcastChannel = "paper"
	
	// Test-friendly logging
	config.Logging.Level = "debug"
	config.Logging.Format = "console"
	
	return config
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	// Expand home directory
	if configPath[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, configPath[1:])
	}
	
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// If config file doesn't exist, create default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		if err := SaveConfig(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return config, nil
	}
	
	// Load existing config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	// Expand home directory
	if configPath[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, configPath[1:])
	}
	
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	// Check environment variable first
	if configPath := os.Getenv("ALGONIUS_WALLET_CONFIG"); configPath != "" {
		return configPath
	}
	
	// Default path
	return "~/.algonius-wallet/config.yaml"
}

// LoadConfigWithFallback loads config with test fallback based on RUN_MODE
func LoadConfigWithFallback(logger *zap.Logger) (*Config, error) {
	runMode := os.Getenv("RUN_MODE")
	
	if runMode == "test" {
		if logger != nil {
			logger.Info("Using test configuration (RUN_MODE=test)")
		}
		return TestConfig(), nil
	}
	
	configPath := GetConfigPath()
	config, err := LoadConfig(configPath)
	if err != nil {
		if logger != nil {
			logger.Warn("Failed to load config, using defaults", 
				zap.String("config_path", configPath),
				zap.Error(err))
		}
		return DefaultConfig(), nil
	}
	
	if logger != nil {
		logger.Info("Configuration loaded successfully", 
			zap.String("config_path", configPath),
			zap.String("network_mode", config.Wallet.NetworkMode))
	}
	
	return config, nil
}