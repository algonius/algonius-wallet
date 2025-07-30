// SPDX-License-Identifier: Apache-2.0
package chain

import (
	"time"
)

// SolanaConfig holds configuration for Solana chain operations
type SolanaConfig struct {
	// Private key in base58 format (only for non-test mode)
	PrivateKey string `mapstructure:"private_key" json:"private_key,omitempty"`
	
	// RPC endpoints with backup support
	RPCEndpoints []string `mapstructure:"rpc_endpoints" json:"rpc_endpoints"`
	
	// WebSocket endpoint for real-time updates
	WSEndpoint string `mapstructure:"ws_endpoint" json:"ws_endpoint"`
	
	// Transaction commitment level
	Commitment string `mapstructure:"commitment" json:"commitment"`
	
	// Reserve SOL amount for gas fees
	ReserveSOL float64 `mapstructure:"reserve_sol" json:"reserve_sol"`
	
	// Retry configuration
	Retry RetryConfig `mapstructure:"retry" json:"retry"`
	
	// Confirmation options
	Confirmation ConfirmationConfig `mapstructure:"confirmation" json:"confirmation"`
	
	// Jito MEV tips configuration
	JitoConfig JitoConfig `mapstructure:"jito" json:"jito"`
}

// RetryConfig defines retry behavior for failed transactions
type RetryConfig struct {
	// Maximum number of retry attempts
	MaxRetries int `mapstructure:"max_retries" json:"max_retries"`
	
	// Slippage increment per retry attempt (in basis points)
	SlippageIncrementBps int `mapstructure:"slippage_increment_bps" json:"slippage_increment_bps"`
	
	// Base delay between retry attempts
	BaseRetryDelay time.Duration `mapstructure:"base_retry_delay" json:"base_retry_delay"`
	
	// Maximum total slippage allowed (in basis points)
	MaxTotalSlippageBps int `mapstructure:"max_total_slippage_bps" json:"max_total_slippage_bps"`
	
	// Gas level strategy for retries
	GasStrategy string `mapstructure:"gas_strategy" json:"gas_strategy"`
}

// ConfirmationConfig defines transaction confirmation behavior
type ConfirmationConfig struct {
	// Timeout for transaction confirmation
	Timeout time.Duration `mapstructure:"timeout" json:"timeout"`
	
	// Polling interval for checking confirmation status
	PollInterval time.Duration `mapstructure:"poll_interval" json:"poll_interval"`
	
	// Required confirmation level
	RequiredConfirmations int `mapstructure:"required_confirmations" json:"required_confirmations"`
}

// JitoConfig defines MEV protection settings
type JitoConfig struct {
	// Whether to use Jito for MEV protection
	Enabled bool `mapstructure:"enabled" json:"enabled"`
	
	// Base tip amount in lamports
	BaseTipLamports uint64 `mapstructure:"base_tip_lamports" json:"base_tip_lamports"`
	
	// Maximum tip amount in lamports
	MaxTipLamports uint64 `mapstructure:"max_tip_lamports" json:"max_tip_lamports"`
	
	// Tip escalation strategy
	TipStrategy string `mapstructure:"tip_strategy" json:"tip_strategy"`
}

// DefaultSolanaConfig returns default configuration for Solana
func DefaultSolanaConfig() *SolanaConfig {
	return &SolanaConfig{
		RPCEndpoints: []string{
			"https://api.mainnet-beta.solana.com",
			"https://solana-api.projectserum.com",
		},
		WSEndpoint: "wss://api.mainnet-beta.solana.com",
		Commitment: "confirmed",
		ReserveSOL: 0.01, // Reserve 0.01 SOL for gas
		Retry: RetryConfig{
			MaxRetries:           3,
			SlippageIncrementBps: 50, // 0.5% per retry
			BaseRetryDelay:       time.Second * 2,
			MaxTotalSlippageBps:  1000, // 10% max slippage
			GasStrategy:          "level_up",
		},
		Confirmation: ConfirmationConfig{
			Timeout:               time.Minute * 2,
			PollInterval:          time.Second * 3,
			RequiredConfirmations: 1,
		},
		JitoConfig: JitoConfig{
			Enabled:         false, // Disabled by default
			BaseTipLamports: 1000,  // 0.000001 SOL
			MaxTipLamports:  100000, // 0.0001 SOL
			TipStrategy:     "exponential",
		},
	}
}

// TestSolanaConfig returns configuration suitable for testing
func TestSolanaConfig() *SolanaConfig {
	config := DefaultSolanaConfig()
	
	// Use devnet for testing
	config.RPCEndpoints = []string{
		"https://api.devnet.solana.com",
	}
	config.WSEndpoint = "wss://api.devnet.solana.com"
	
	// Faster confirmations for testing
	config.Confirmation.Timeout = time.Second * 30
	config.Confirmation.PollInterval = time.Second * 1
	
	// Less aggressive retry for testing
	config.Retry.MaxRetries = 1
	config.Retry.BaseRetryDelay = time.Millisecond * 500
	
	return config
}