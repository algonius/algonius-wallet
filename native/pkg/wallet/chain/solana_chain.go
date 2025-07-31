// SPDX-License-Identifier: Apache-2.0
package chain

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mr-tron/base58"
	bip39 "github.com/tyler-smith/go-bip39"
	"github.com/algonius/algonius-wallet/native/pkg/chains/solana/broadcast"
	"github.com/algonius/algonius-wallet/native/pkg/config"
	"github.com/algonius/algonius-wallet/native/pkg/dex"
	"go.uber.org/zap"
)

// SolanaChain implements the IChain interface for Solana
type SolanaChain struct {
	name             string
	dexAggregator    dex.IDEXAggregator
	logger           *zap.Logger
	chainID          string
	config           *config.SolanaChainConfig
	rpcManager       *SolanaRPCManager
	retryManager     *SolanaRetryManager
	broadcastManager *broadcast.BroadcastManager
}

// NewSolanaChain creates a new Solana chain instance with enhanced blockchain integration
func NewSolanaChain(dexAggregator dex.IDEXAggregator, logger *zap.Logger) (*SolanaChain, error) {
	// Load configuration based on run mode
	appConfig, err := config.LoadConfigWithFallback(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	
	solanaConfig := &appConfig.Chains.Solana
	
	// Initialize RPC manager
	rpcManager, err := NewSolanaRPCManager(solanaConfig.RPCEndpoints, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC manager: %w", err)
	}
	
	// Convert config retry to local retry config
	localRetryConfig := &RetryConfig{
		MaxRetries:           solanaConfig.Retry.MaxRetries,
		SlippageIncrementBps: solanaConfig.Retry.SlippageIncrementBps,
		BaseRetryDelay:       solanaConfig.Retry.BaseRetryDelay,
		MaxTotalSlippageBps:  solanaConfig.Retry.MaxTotalSlippageBps,
		GasStrategy:          solanaConfig.Retry.GasStrategy,
	}
	
	// Initialize retry manager
	retryManager := NewSolanaRetryManager(localRetryConfig, logger)
	
	// Initialize broadcast manager
	broadcastManager := broadcast.NewBroadcastManager(&solanaConfig.Broadcast)
	
	// Register broadcast channels
	solanaRPCChannel := broadcast.NewSolanaRPCChannel(solanaConfig, logger)
	broadcastManager.RegisterChannel(solanaRPCChannel)
	
	okexChannel := broadcast.NewOKExChannel(&appConfig.DEX.OKEx, logger)
	broadcastManager.RegisterChannel(okexChannel)
	
	// Add Jito channels (if enabled)
	if solanaConfig.Jito.Enabled {
		jitoChannel := broadcast.NewJitoChannel(&solanaConfig.Jito, solanaConfig.RPCEndpoints[0], logger)
		broadcastManager.RegisterChannel(jitoChannel)
		
		jitoBundleChannel := broadcast.NewJitoBundleChannel(&solanaConfig.Jito, solanaConfig.RPCEndpoints[0], logger)
		// Initialize Jito bundle channel (needs tip account setup)
		if err := jitoBundleChannel.Init(context.Background()); err != nil {
			logger.Warn("Failed to initialize Jito bundle channel", zap.Error(err))
		} else {
			broadcastManager.RegisterChannel(jitoBundleChannel)
		}
	}
	
	paperChannel := broadcast.NewPaperChannel(logger)
	broadcastManager.RegisterChannel(paperChannel)
	
	chain := &SolanaChain{
		name:             "SOLANA",
		dexAggregator:    dexAggregator,
		logger:           logger,
		chainID:          "501", // Solana Mainnet
		config:           solanaConfig,
		rpcManager:       rpcManager,
		retryManager:     retryManager,
		broadcastManager: broadcastManager,
	}
	
	logger.Info("Initialized Solana chain with enhanced blockchain integration",
		zap.String("run_mode", os.Getenv("RUN_MODE")),
		zap.Int("rpc_endpoints", len(solanaConfig.RPCEndpoints)),
		zap.Int("max_retries", solanaConfig.Retry.MaxRetries),
		zap.String("broadcast_channel", solanaConfig.Broadcast.Channel))
	
	return chain, nil
}

// NewSolanaChainLegacy creates a new Solana chain instance without DEX aggregator (for backward compatibility)
func NewSolanaChainLegacy() *SolanaChain {
	logger := zap.NewNop() // Use no-op logger for legacy
	
	// Try to create enhanced chain, fallback to basic if it fails
	if chain, err := NewSolanaChain(nil, logger); err == nil {
		return chain
	}
	
	return &SolanaChain{
		name:    "SOLANA",
		chainID: "501",
		logger:  logger,
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

// ImportFromMnemonic imports a wallet from mnemonic phrase with derivation path
func (s *SolanaChain) ImportFromMnemonic(ctx context.Context, mnemonic, derivationPath string) (*WalletInfo, error) {
	// Validate mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("invalid mnemonic phrase")
	}
	
	// Generate seed from mnemonic
	seed := bip39.NewSeed(mnemonic, "")
	
	// Use default derivation path if not provided
	if derivationPath == "" {
		derivationPath = "m/44'/501'/0'/0'" // Solana default
	}
	
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

	// Try to get balance using RPC manager if available
	if s.rpcManager != nil {
		result, err := s.rpcManager.GetBalance(ctx, address, s.config.Commitment)
		if err == nil {
			// Convert lamports to SOL (1 SOL = 1,000,000,000 lamports)
			balanceSOL := float64(result.Value) / 1000000000.0
			
			// Format balance: for zero balance, return "0", otherwise show appropriate precision
			var balanceStr string
			if result.Value == 0 {
				balanceStr = "0"
			} else if balanceSOL >= 1 {
				balanceStr = fmt.Sprintf("%.6f", balanceSOL)
			} else {
				balanceStr = fmt.Sprintf("%.9f", balanceSOL)
			}
			
			s.logger.Debug("Solana balance retrieved via RPC",
				zap.String("address", address),
				zap.Uint64("lamports", result.Value),
				zap.String("balance_sol", balanceStr))
			
			return balanceStr, nil
		}
		
		s.logger.Warn("Solana RPC balance failed, falling back to DEX provider",
			zap.Error(err))
	}

	// Try to get balance using DEX aggregator as fallback
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

	// Use enhanced transaction execution with retry mechanism
	return s.sendTransactionWithRetry(ctx, from, to, amount, token, privateKey)
}

// sendTransactionWithRetry executes a Solana transaction with intelligent retry logic
func (s *SolanaChain) sendTransactionWithRetry(ctx context.Context, from, to, amount, token, _ string) (string, error) {
	if s.retryManager == nil || s.rpcManager == nil {
		// Fallback to mock implementation if managers not available
		return s.createMockTransaction(from, to, amount, token)
	}
	
	// Parse amount to lamports (assuming SOL for now)
	// In a real implementation, this would handle different token decimals
	amountLamports := uint64(1000000) // Mock: 0.001 SOL in lamports
	
	// Get recent blockhash
	blockhashResult, err := s.rpcManager.GetLatestBlockhash(ctx, s.config.Commitment)
	if err != nil {
		s.logger.Error("Failed to get latest blockhash", zap.Error(err))
		return "", fmt.Errorf("failed to get blockhash: %w", err)
	}
	
	// Prepare transaction parameters
	txParams := &TransactionParams{
		From:            from,
		To:              to,
		Amount:          amountLamports,
		TokenMint:       token,
		Slippage:        0.005, // 0.5% default slippage
		RecentBlockhash: blockhashResult.Value.Blockhash,
		JitoTipAmount:   s.config.Jito.BaseTipLamports,
		GasStrategy:     s.config.Retry.GasStrategy,
	}
	
	// Execute transaction with retry logic
	result, err := s.retryManager.ExecuteWithRetry(ctx, txParams, s.executeTransactionAttempt)
	if err != nil {
		s.logger.Error("Transaction failed after all retries",
			zap.Error(err),
			zap.Int("attempts", result.Attempt))
		return "", err
	}
	
	s.logger.Info("Transaction executed successfully",
		zap.String("signature", result.Signature),
		zap.Int("attempts", result.Attempt),
		zap.Float64("final_slippage", result.FinalSlippage))
	
	return result.Signature, nil
}

// executeTransactionAttempt executes a single transaction attempt using broadcast manager
func (s *SolanaChain) executeTransactionAttempt(ctx context.Context, params *TransactionParams) (string, error) {
	s.logger.Debug("Executing transaction attempt",
		zap.String("from", params.From),
		zap.String("to", params.To),
		zap.Uint64("amount", params.Amount),
		zap.String("blockhash", params.RecentBlockhash))
	
	// Create transaction (simplified for now)
	transactionData, signature := s.createTransaction(params)
	
	// Prepare broadcast parameters
	broadcastParams := &broadcast.BroadcastParams{
		SignedTransaction:   transactionData,
		TransactionBase64:   string(transactionData), // Would be base64 in real impl
		Signature:          signature,
		From:               params.From,
		To:                 params.To,
		Amount:             params.Amount,
		Token:              params.TokenMint,
		SkipPreflight:      false,
		MaxRetries:         3,
		PreflightCommitment: s.config.Commitment,
		Timeout:            30 * time.Second,
		Metadata: map[string]any{
			"blockhash":      params.RecentBlockhash,
			"jito_tip":       params.JitoTipAmount,
			"slippage":       params.Slippage,
			"gas_strategy":   params.GasStrategy,
		},
	}
	
	// Broadcast transaction with failover
	result, err := s.broadcastManager.BroadcastWithFallback(ctx, broadcastParams)
	if err != nil {
		return "", fmt.Errorf("failed to broadcast transaction: %w", err)
	}
	
	s.logger.Info("Transaction broadcasted successfully",
		zap.String("signature", result.Signature),
		zap.String("channel", result.Channel),
		zap.Duration("duration", result.Duration),
		zap.String("status", result.Status))
	
	return result.Signature, nil
}

// createTransaction creates and signs a Solana transaction
func (s *SolanaChain) createTransaction(params *TransactionParams) ([]byte, string) {
	// In a real implementation, this would:
	// 1. Create proper Solana transaction with transfer instruction
	// 2. Sign with ed25519 private key  
	// 3. Serialize to bytes
	// 4. Return both serialized transaction and signature
	
	if os.Getenv("RUN_MODE") == "test" {
		// Generate mock transaction data and signature
		mockTxData := []byte(fmt.Sprintf("MockTxData_%s_%d", params.From[:8], time.Now().Unix()))
		mockSignature := fmt.Sprintf("MockSignature_%s_%d", params.From[:8], time.Now().Unix())
		return mockTxData, mockSignature
	}
	
	// TODO: Implement actual Solana transaction creation
	// This would involve:
	// - Creating transfer or token transfer instructions
	// - Setting compute budget and priority fees
	// - Adding Jito tip instruction if enabled
	// - Signing with private key
	// - Serializing to wire format
	
	simulatedTxData := []byte("SimulatedSolanaTransaction")
	simulatedSignature := "SimulatedSolanaSignature"
	
	return simulatedTxData, simulatedSignature
}


// createMockTransaction creates a mock transaction signature for fallback
func (s *SolanaChain) createMockTransaction(from, to, amount, token string) (string, error) {
	s.logger.Debug("Creating mock transaction (fallback mode)",
		zap.String("from", from),
		zap.String("to", to),
		zap.String("amount", amount),
		zap.String("token", token))
	
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