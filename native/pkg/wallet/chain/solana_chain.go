// SPDX-License-Identifier: Apache-2.0
package chain

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mr-tron/base58"
	bip39 "github.com/tyler-smith/go-bip39"
	"github.com/algonius/algonius-wallet/native/pkg/dex"
	"go.uber.org/zap"
)

// SolanaChain implements the IChain interface for Solana
type SolanaChain struct {
	name         string
	dexAggregator dex.IDEXAggregator
	logger       *zap.Logger
	chainID      string
}

// NewSolanaChain creates a new Solana chain instance
func NewSolanaChain(dexAggregator dex.IDEXAggregator, logger *zap.Logger) *SolanaChain {
	return &SolanaChain{
		name:         "SOLANA",
		dexAggregator: dexAggregator,
		logger:       logger,
		chainID:      "501", // Solana Mainnet
	}
}

// NewSolanaChainLegacy creates a new Solana chain instance without DEX aggregator (for backward compatibility)
func NewSolanaChainLegacy() *SolanaChain {
	return &SolanaChain{
		name:    "SOLANA",
		chainID: "501",
	}
}

// GetChainName returns the name of the chain
func (s *SolanaChain) GetChainName() string {
	return s.name
}

// CreateWallet generates a new Solana wallet
func (s *SolanaChain) CreateWallet(ctx context.Context) (*WalletInfo, error) {
	// Generate entropy for mnemonic
	entropy, err := bip39.NewEntropy(128) // 128 bits = 12 words
	if err != nil {
		return nil, fmt.Errorf("failed to generate entropy: %w", err)
	}

	// Generate mnemonic phrase
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	// Generate seed from mnemonic
	seed := bip39.NewSeed(mnemonic, "")

	// For Solana, we need a 32-byte seed for the ed25519 keypair
	// Truncate or pad to 32 bytes if needed
	keySeed := make([]byte, 32)
	copy(keySeed, seed)

	// Generate ed25519 keypair
	publicKey, privateKey, err := ed25519.GenerateKey(strings.NewReader(string(keySeed)))
	if err != nil {
		return nil, fmt.Errorf("failed to generate ed25519 keypair: %w", err)
	}

	// Solana addresses are base58-encoded public keys
	address := base58.Encode(publicKey)

	// Convert keys to base58 strings for Solana
	privateKeyB58 := base58.Encode(privateKey.Seed())
	publicKeyB58 := base58.Encode(publicKey)

	return &WalletInfo{
		Address:    address,
		PublicKey:  publicKeyB58,
		PrivateKey: privateKeyB58,
		Mnemonic:   mnemonic,
	}, nil
}

// GetBalance retrieves the balance for a Solana address
func (s *SolanaChain) GetBalance(ctx context.Context, address string, token string) (string, error) {
	// Validate address format (base58)
	if address == "" {
		return "", errors.New("invalid Solana address format")
	}

	// Try to decode to check if it's valid base58
	_, err := base58.Decode(address)
	if err != nil {
		return "", errors.New("invalid Solana address format")
	}

	// Normalize token name
	token = strings.ToUpper(strings.TrimSpace(token))
	if token == "" {
		token = "SOL"
	}

	// For now, only support SOL (native token)
	if token != "SOL" {
		return "", fmt.Errorf("unsupported token: %s", token)
	}

	// Try to get balance using DEX aggregator if available
	if s.dexAggregator != nil {
		providers := s.dexAggregator.GetSupportedProviders(s.chainID)
		if len(providers) > 0 {
			// Try first available provider
			provider, err := s.dexAggregator.GetProviderByName(providers[0])
			if err == nil {
				balanceInfo, err := provider.GetBalance(ctx, address, token, s.chainID)
				if err == nil {
					s.logger.Debug("Solana balance retrieved via DEX provider",
						zap.String("provider", providers[0]),
						zap.String("balance", balanceInfo.Balance))
					return balanceInfo.Balance, nil
				}
				s.logger.Warn("Solana DEX provider balance failed, falling back to mock",
					zap.String("provider", providers[0]),
					zap.Error(err))
			}
		}
	}

	// TODO: Implement actual balance retrieval from Solana RPC
	// This is a mock implementation that returns "0"
	// In a real implementation, you would:
	// 1. Connect to a Solana RPC endpoint
	// 2. Use getBalance RPC method to get the balance
	// 3. Convert from lamports to SOL (1 SOL = 1,000,000,000 lamports)
	return "0", nil
}

// SendTransaction sends a transaction on the Solana network
func (s *SolanaChain) SendTransaction(ctx context.Context, from, to string, amount string, token string, privateKey string) (string, error) {
	// Validate addresses
	if from == "" || to == "" {
		return "", errors.New("from and to addresses are required")
	}

	// Validate base58 format
	if _, err := base58.Decode(from); err != nil {
		return "", errors.New("invalid from address format")
	}
	if _, err := base58.Decode(to); err != nil {
		return "", errors.New("invalid to address format")
	}

	// Validate amount is not empty
	if amount == "" {
		return "", errors.New("amount cannot be empty")
	}

	// Validate private key format
	if privateKey == "" {
		return "", errors.New("private key is required")
	}

	// Try to decode private key
	if _, err := base58.Decode(privateKey); err != nil {
		return "", errors.New("invalid private key format")
	}

	// Normalize token - empty means SOL
	token = strings.TrimSpace(token)
	if token == "" {
		token = "SOL"
	}

	// Handle token validation
	// For Solana, token addresses are base58-encoded public keys (program addresses)
	if strings.ToUpper(token) != "SOL" {
		if _, err := base58.Decode(token); err != nil {
			return "", fmt.Errorf("invalid token program address: %s", token)
		}
	}

	// Additional security checks
	if from == to {
		return "", errors.New("cannot send to the same address")
	}

	// Try to execute swap using DEX aggregator if it's a token swap
	if s.dexAggregator != nil && strings.ToUpper(token) != "SOL" {
		swapParams := dex.SwapParams{
			FromToken:    "SOL",
			ToToken:      token,
			Amount:       amount,
			Slippage:     0.005, // 0.5% default slippage
			FromAddress:  from,
			ToAddress:    to,
			ChainID:      s.chainID,
			PrivateKey:   privateKey,
		}

		// Try to get best quote and execute swap
		quote, err := s.dexAggregator.GetBestQuote(ctx, swapParams)
		if err == nil {
			s.logger.Info("Executing Solana token swap via DEX aggregator",
				zap.String("provider", quote.Provider),
				zap.String("fromAmount", quote.FromAmount),
				zap.String("toAmount", quote.ToAmount))

			result, err := s.dexAggregator.ExecuteSwapWithProvider(ctx, quote.Provider, swapParams)
			if err == nil {
				return result.TxHash, nil
			}
			s.logger.Warn("Solana DEX swap failed, falling back to direct transfer",
				zap.Error(err))
		} else {
			s.logger.Debug("No Solana DEX quote available, proceeding with direct transfer",
				zap.Error(err))
		}
	}

	// TODO: Implement actual Solana transaction creation and signing
	// This is an enhanced mock implementation with proper validation
	// In a real implementation, you would:
	// 1. Parse amount to lamports (for SOL) or token decimals (for SPL tokens)
	// 2. Get recent blockhash from Solana RPC
	// 3. Create a transaction with appropriate instructions
	// 4. For SPL tokens: Create token transfer instruction
	// 5. Sign the transaction with the private key
	// 6. Broadcast the transaction to the Solana network
	// 7. Return the actual transaction signature

	// For demo, we'll just create a mock signature
	signatureBytes := make([]byte, 64)
	for i := range signatureBytes {
		signatureBytes[i] = byte(i % 256)
	}
	signature := base58.Encode(signatureBytes)
	
	return signature, nil
}

// EstimateGas estimates gas requirements for a Solana transaction
// Note: Solana uses "compute units" instead of gas
func (s *SolanaChain) EstimateGas(ctx context.Context, from, to string, amount string, token string) (gasLimit uint64, gasPrice string, err error) {
	// Validate addresses
	if from == "" || to == "" {
		return 0, "", errors.New("from and to addresses are required")
	}

	// Validate base58 format
	if _, err := base58.Decode(from); err != nil {
		return 0, "", errors.New("invalid from address format")
	}
	if _, err := base58.Decode(to); err != nil {
		return 0, "", errors.New("invalid to address format")
	}

	// Normalize token - empty means SOL
	token = strings.TrimSpace(token)
	if token == "" {
		token = "SOL"
	}

	// Basic compute unit estimation based on transaction type for Solana
	var baseComputeUnits uint64
	var baseComputePrice string = "1" // 1 microlamport per compute unit as default

	if strings.ToUpper(token) == "SOL" {
		// Simple SOL transfer
		baseComputeUnits = 150
	} else {
		// SPL token transfer requires more compute units
		if _, err := base58.Decode(token); err != nil {
			return 0, "", fmt.Errorf("invalid token program address: %s", token)
		}
		baseComputeUnits = 5000 // Typical compute units for SPL token transfer
	}

	// Try to get gas estimate from DEX aggregator if available
	if s.dexAggregator != nil {
		swapParams := dex.SwapParams{
			FromToken:   "SOL",
			ToToken:     token,
			Amount:      amount,
			FromAddress: from,
			ToAddress:   to,
			ChainID:     s.chainID,
		}

		providers := s.dexAggregator.GetSupportedProviders(s.chainID)
		if len(providers) > 0 {
			provider, err := s.dexAggregator.GetProviderByName(providers[0])
			if err == nil {
				gasLimit, gasPrice, err := provider.EstimateGas(ctx, swapParams)
				if err == nil {
					s.logger.Debug("Solana compute units estimate from DEX provider",
						zap.String("provider", providers[0]),
						zap.Uint64("computeUnits", gasLimit),
						zap.String("computePrice", gasPrice))
					return gasLimit, gasPrice, nil
				}
			}
		}
	}

	// TODO: In a real implementation, you would:
	// 1. Connect to a Solana RPC endpoint
	// 2. Use simulateTransaction RPC method to get actual compute unit estimation
	// 3. Get current compute unit price from the network
	// 4. Apply safety multipliers

	return baseComputeUnits, baseComputePrice, nil
}

// ConfirmTransaction checks the confirmation status of a Solana transaction
func (s *SolanaChain) ConfirmTransaction(ctx context.Context, txHash string, requiredConfirmations uint64) (*TransactionConfirmation, error) {
	// Validate transaction signature format (base58)
	if txHash == "" {
		return nil, errors.New("transaction signature cannot be empty")
	}

	// Validate base58 format
	signatureBytes, err := base58.Decode(txHash)
	if err != nil {
		return nil, errors.New("invalid transaction signature format")
	}

	// Validate length (64 bytes for Solana signatures)
	if len(signatureBytes) != 64 {
		return nil, errors.New("invalid transaction signature length")
	}

	// Set default required confirmations if not provided
	if requiredConfirmations == 0 {
		requiredConfirmations = 1 // Default for Solana (single confirmation is typically sufficient)
	}

	// TODO: Implement actual transaction confirmation checking for Solana
	// This is a mock implementation for development purposes
	// In a real implementation, you would:
	// 1. Connect to a Solana RPC endpoint
	// 2. Use getSignatureStatuses RPC method to get the transaction status
	// 3. Get block details for timestamp using getBlock RPC method
	// 4. Return actual transaction details

	// For development purposes, simulate different transaction states
	// This will be replaced with actual blockchain queries
	var status string
	var confirmations uint64
	var blockNumber uint64 = 200000000 // Mock Solana block number (much higher than ETH/BSC)
	var gasUsed string = "150"         // Standard SOL transfer compute units
	var transactionFee string = "0.000005" // Mock fee (typical Solana transaction fee)
	var timestamp time.Time = time.Now().Add(-1 * time.Minute) // Mock timestamp (Solana is very fast)

	// Simple logic to simulate transaction states based on signature
	lastByte := signatureBytes[len(signatureBytes)-1]

	switch {
	case lastByte%15 == 0: // ~6.7% chance of failed transaction
		status = "failed"
		confirmations = 0
	case lastByte%5 == 0: // 20% chance of pending transaction
		status = "pending"
		confirmations = 0 // Random confirmations < required
	default: // ~73.3% chance of confirmed transaction (Solana confirms very fast)
		status = "confirmed"
		confirmations = requiredConfirmations
	}

	// Mock different compute units for different transaction types
	if strings.Contains(strings.ToLower(txHash), "spl") || lastByte%4 == 1 {
		gasUsed = "5000" // SPL token transfer compute units
		transactionFee = "0.000015" // Higher fee for token transfers
	}

	return &TransactionConfirmation{
		Status:                status,
		Confirmations:         confirmations,
		RequiredConfirmations: requiredConfirmations,
		BlockNumber:           blockNumber,
		GasUsed:               gasUsed,
		TransactionFee:        transactionFee,
		Timestamp:             timestamp,
		TxHash:                txHash,
	}, nil
}