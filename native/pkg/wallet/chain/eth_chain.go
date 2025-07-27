// SPDX-License-Identifier: Apache-2.0
package chain

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	bip39 "github.com/tyler-smith/go-bip39"
	"github.com/algonius/algonius-wallet/native/pkg/dex"
	"go.uber.org/zap"
)

// ETHChain implements the IChain interface for Ethereum
type ETHChain struct {
	name         string
	dexAggregator dex.IDEXAggregator
	logger       *zap.Logger
	chainID      string
}

// NewETHChain creates a new ETH chain instance
func NewETHChain(dexAggregator dex.IDEXAggregator, logger *zap.Logger) *ETHChain {
	return &ETHChain{
		name:         "ETH",
		dexAggregator: dexAggregator,
		logger:       logger,
		chainID:      "1", // Ethereum Mainnet
	}
}

// NewETHChainLegacy creates a new ETH chain instance without DEX aggregator (for backward compatibility)
func NewETHChainLegacy() *ETHChain {
	return &ETHChain{
		name:    "ETH",
		chainID: "1",
	}
}

// GetChainName returns the name of the chain
func (e *ETHChain) GetChainName() string {
	return e.name
}

// CreateWallet generates a new Ethereum wallet
func (e *ETHChain) CreateWallet(ctx context.Context) (*WalletInfo, error) {
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

	// For simplicity, we'll use the first 32 bytes as private key
	// In a production system, you'd want to use proper HD wallet derivation
	if len(seed) < 32 {
		return nil, errors.New("insufficient seed length")
	}

	// Create private key from seed
	privateKey, err := crypto.ToECDSA(seed[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create private key: %w", err)
	}

	// Derive public key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("failed to get public key")
	}

	// Generate Ethereum address
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Convert keys to hex strings
	privateKeyHex := hexutil.Encode(crypto.FromECDSA(privateKey))
	publicKeyHex := hexutil.Encode(crypto.FromECDSAPub(publicKeyECDSA))

	return &WalletInfo{
		Address:    address.Hex(),
		PublicKey:  publicKeyHex,
		PrivateKey: privateKeyHex,
		Mnemonic:   mnemonic,
	}, nil
}

// GetBalance retrieves the balance for an Ethereum address
func (e *ETHChain) GetBalance(ctx context.Context, address string, token string) (string, error) {
	// Validate address format
	if !common.IsHexAddress(address) {
		return "", errors.New("invalid Ethereum address format")
	}

	// Normalize token name
	token = strings.ToUpper(strings.TrimSpace(token))
	if token == "" {
		token = "ETH"
	}

	// Support standardized token identifiers
	// ETH can be identified as "ETH" or "ETHER"
	supportedTokens := map[string]bool{
		"ETH":   true,
		"ETHER": true,
	}

	if !supportedTokens[token] {
		// Check if it's a contract address for ERC-20 tokens
		if !common.IsHexAddress(token) {
			return "", fmt.Errorf("unsupported token: %s", token)
		}
		// For now, we'll treat any valid hex address as a potential ERC-20 token
		// TODO: In a real implementation, verify it's a valid ERC-20 contract
	}

	// Try to get balance using DEX aggregator if available
	if e.dexAggregator != nil {
		providers := e.dexAggregator.GetSupportedProviders(e.chainID)
		if len(providers) > 0 {
			// Try first available provider
			provider, err := e.dexAggregator.GetProviderByName(providers[0])
			if err == nil {
				balanceInfo, err := provider.GetBalance(ctx, address, token, e.chainID)
				if err == nil {
					e.logger.Debug("Balance retrieved via DEX provider",
						zap.String("provider", providers[0]),
						zap.String("balance", balanceInfo.Balance))
					return balanceInfo.Balance, nil
				}
				e.logger.Warn("DEX provider balance failed, falling back to mock",
					zap.String("provider", providers[0]),
					zap.Error(err))
			}
		}
	}

	// TODO: Implement actual balance retrieval from Ethereum node
	// This is a mock implementation that returns "0"
	// In a real implementation, you would:
	// 1. Connect to an Ethereum node (Infura, Alchemy, etc.)
	// 2. For ETH: Use ethclient.BalanceAt() to get the balance
	// 3. For ERC-20: Use the contract's balanceOf function
	// 4. Convert from Wei to Ether or token decimals
	return "0", nil
}

// SendTransaction sends a transaction on the Ethereum network
func (e *ETHChain) SendTransaction(ctx context.Context, from, to string, amount string, token string, privateKey string) (string, error) {
	// Validate addresses
	if !common.IsHexAddress(from) {
		return "", errors.New("invalid from address format")
	}
	if !common.IsHexAddress(to) {
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

	// Validate private key is valid hex
	if !strings.HasPrefix(privateKey, "0x") {
		return "", errors.New("private key must be in hex format (0x...)")
	}

	// Normalize token - empty means ETH
	token = strings.TrimSpace(token)
	if token == "" {
		token = "ETH"
	}

	// Handle token validation
	isERC20 := false
	if strings.ToUpper(token) != "ETH" {
		// Check if it's a valid contract address for ERC-20 token
		if !common.IsHexAddress(token) {
			return "", fmt.Errorf("invalid token contract address: %s", token)
		}
		isERC20 = true
	}

	// Additional security checks
	fromAddr := common.HexToAddress(from)
	toAddr := common.HexToAddress(to)

	// Prevent sending to zero address
	if toAddr == (common.Address{}) {
		return "", errors.New("cannot send to zero address")
	}

	// Prevent sending to same address (unless explicitly allowed)
	if fromAddr == toAddr {
		return "", errors.New("cannot send to the same address")
	}

	// Try to execute swap using DEX aggregator if it's a token swap
	if e.dexAggregator != nil && isERC20 {
		swapParams := dex.SwapParams{
			FromToken:    "ETH",
			ToToken:      token,
			Amount:       amount,
			Slippage:     0.005, // 0.5% default slippage
			FromAddress:  from,
			ToAddress:    to,
			ChainID:      e.chainID,
			PrivateKey:   privateKey,
		}

		// Try to get best quote and execute swap
		quote, err := e.dexAggregator.GetBestQuote(ctx, swapParams)
		if err == nil {
			e.logger.Info("Executing token swap via DEX aggregator",
				zap.String("provider", quote.Provider),
				zap.String("fromAmount", quote.FromAmount),
				zap.String("toAmount", quote.ToAmount))

			result, err := e.dexAggregator.ExecuteSwapWithProvider(ctx, quote.Provider, swapParams)
			if err == nil {
				return result.TxHash, nil
			}
			e.logger.Warn("DEX swap failed, falling back to direct transfer",
				zap.Error(err))
		} else {
			e.logger.Debug("No DEX quote available, proceeding with direct transfer",
				zap.Error(err))
		}
	}

	// TODO: Implement actual transaction creation and signing
	// This is an enhanced mock implementation with proper validation
	// In a real implementation, you would:
	// 1. Parse amount to Wei (for ETH) or token decimals (for ERC-20)
	// 2. Get current nonce for the from address
	// 3. Estimate gas for the transaction
	// 4. For ERC-20: Create contract call data for transfer function
	// 5. Create transaction with proper gas price and gas limit
	// 6. Sign the transaction with the private key
	// 7. Broadcast the transaction to the network
	// 8. Return the actual transaction hash

	// Generate a realistic-looking transaction hash for demo purposes
	var hashInput string
	if isERC20 {
		hashInput = fmt.Sprintf("ETH-ERC20-%s-%s%s%s", token, from, to, amount)
	} else {
		hashInput = fmt.Sprintf("ETH-%s%s%s", from, to, amount)
	}
	hash := crypto.Keccak256Hash([]byte(hashInput))
	return hash.Hex(), nil
}

// EstimateGas estimates gas requirements for an Ethereum transaction
func (e *ETHChain) EstimateGas(ctx context.Context, from, to string, amount string, token string) (gasLimit uint64, gasPrice string, err error) {
	// Validate addresses
	if !common.IsHexAddress(from) {
		return 0, "", errors.New("invalid from address format")
	}
	if !common.IsHexAddress(to) {
		return 0, "", errors.New("invalid to address format")
	}

	// Normalize token - empty means ETH
	token = strings.TrimSpace(token)
	if token == "" {
		token = "ETH"
	}

	// Basic gas estimation based on transaction type
	var baseGasLimit uint64
	var baseGasPrice string = "20" // 20 gwei as default

	if strings.ToUpper(token) == "ETH" {
		// Simple ETH transfer
		baseGasLimit = 21000
	} else {
		// ERC-20 token transfer requires more gas
		if !common.IsHexAddress(token) {
			return 0, "", fmt.Errorf("invalid token contract address: %s", token)
		}
		baseGasLimit = 65000 // Typical gas for ERC-20 transfer
	}

	// Try to get gas estimate from DEX aggregator if available
	if e.dexAggregator != nil {
		swapParams := dex.SwapParams{
			FromToken:   "ETH",
			ToToken:     token,
			Amount:      amount,
			FromAddress: from,
			ToAddress:   to,
			ChainID:     e.chainID,
		}

		providers := e.dexAggregator.GetSupportedProviders(e.chainID)
		if len(providers) > 0 {
			provider, err := e.dexAggregator.GetProviderByName(providers[0])
			if err == nil {
				gasLimit, gasPrice, err := provider.EstimateGas(ctx, swapParams)
				if err == nil {
					e.logger.Debug("Gas estimate from DEX provider",
						zap.String("provider", providers[0]),
						zap.Uint64("gasLimit", gasLimit),
						zap.String("gasPrice", gasPrice))
					return gasLimit, gasPrice, nil
				}
			}
		}
	}

	// TODO: In a real implementation, you would:
	// 1. Connect to an Ethereum node
	// 2. Use eth_estimateGas to get actual gas estimate
	// 3. Get current gas price from the network
	// 4. Apply safety multipliers (e.g., 1.2x for gas limit)
	// 5. Check for network congestion and adjust gas price

	return baseGasLimit, baseGasPrice, nil
}

// ConfirmTransaction checks the confirmation status of an Ethereum transaction
func (e *ETHChain) ConfirmTransaction(ctx context.Context, txHash string, requiredConfirmations uint64) (*TransactionConfirmation, error) {
	// Validate transaction hash format
	if txHash == "" {
		return nil, errors.New("transaction hash cannot be empty")
	}
	
	// Normalize transaction hash
	if !strings.HasPrefix(txHash, "0x") {
		txHash = "0x" + txHash
	}
	
	// Validate hex format and length (32 bytes = 64 hex chars + 0x prefix)
	if len(txHash) != 66 {
		return nil, errors.New("invalid transaction hash length")
	}
	
	// Validate hex format
	_, err := hexutil.Decode(txHash)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction hash format: %w", err)
	}

	// Set default required confirmations if not provided
	if requiredConfirmations == 0 {
		requiredConfirmations = 6 // Default for Ethereum
	}

	// TODO: Implement actual transaction confirmation checking
	// This is a mock implementation for development purposes
	// In a real implementation, you would:
	// 1. Connect to an Ethereum node (ethclient.Dial)
	// 2. Get transaction receipt using client.TransactionReceipt(ctx, common.HexToHash(txHash))
	// 3. Get current block number using client.BlockNumber(ctx)
	// 4. Calculate confirmations = currentBlock - txReceipt.BlockNumber
	// 5. Get block details for timestamp using client.BlockByNumber()
	// 6. Return actual transaction details

	// For development purposes, simulate different transaction states
	// This will be replaced with actual blockchain queries
	var status string
	var confirmations uint64
	var blockNumber uint64 = 18500000 // Mock block number
	var gasUsed string = "21000"      // Standard ETH transfer gas
	var transactionFee string = "0.000420" // Mock fee (20 gwei * 21000 gas)
	var timestamp time.Time = time.Now().Add(-10 * time.Minute) // Mock timestamp

	// Simple logic to simulate transaction states based on hash
	hashBytes := common.HexToHash(txHash).Bytes()
	lastByte := hashBytes[len(hashBytes)-1]
	
	switch {
	case lastByte%10 == 0: // 10% chance of failed transaction
		status = "failed"
		confirmations = 0
	case lastByte%3 == 0: // ~33% chance of pending transaction
		status = "pending" 
		confirmations = uint64(lastByte) % requiredConfirmations // Random confirmations < required
	default: // ~57% chance of confirmed transaction
		status = "confirmed"
		confirmations = requiredConfirmations + uint64(lastByte)%10 // Confirmed with extra confirmations
	}

	// Mock different gas usage for different transaction types
	if strings.Contains(strings.ToLower(txHash), "erc20") || lastByte%4 == 1 {
		gasUsed = "52000" // ERC-20 transfer gas
		transactionFee = "0.001040" // Higher fee for token transfers
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
