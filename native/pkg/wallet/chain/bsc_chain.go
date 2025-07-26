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
)

// BSCChain implements the IChain interface for Binance Smart Chain
type BSCChain struct {
	name string
}

// NewBSCChain creates a new BSC chain instance
func NewBSCChain() *BSCChain {
	return &BSCChain{
		name: "BSC",
	}
}

// GetChainName returns the name of the chain
func (b *BSCChain) GetChainName() string {
	return b.name
}

// CreateWallet generates a new BSC wallet (same as Ethereum since it's EVM-compatible)
func (b *BSCChain) CreateWallet(ctx context.Context) (*WalletInfo, error) {
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

	// Derive the first private key from the seed
	// This is a simplified derivation; in production, you'd use proper BIP-44 derivation
	hash := crypto.Keccak256(seed)
	privateKey, err := crypto.ToECDSA(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to derive private key: %w", err)
	}

	// Get public key
	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("failed to get public key")
	}

	// Generate BSC address (same as Ethereum address format)
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

// GetBalance retrieves the balance for a BSC address
func (b *BSCChain) GetBalance(ctx context.Context, address string, token string) (string, error) {
	// Validate address format
	if !common.IsHexAddress(address) {
		return "", errors.New("invalid BSC address format")
	}

	// Normalize token name
	token = strings.ToUpper(strings.TrimSpace(token))
	if token == "" {
		token = "BNB"
	}

	// Support standardized token identifiers
	// BNB can be identified as "BNB" or "BINANCE"
	supportedTokens := map[string]bool{
		"BNB":     true,
		"BINANCE": true,
	}

	if !supportedTokens[token] {
		// Check if it's a contract address for BEP-20 tokens
		if !common.IsHexAddress(token) {
			return "", fmt.Errorf("unsupported token: %s", token)
		}
		// For now, we'll treat any valid hex address as a potential BEP-20 token
		// TODO: In a real implementation, verify it's a valid BEP-20 contract
	}

	// TODO: Implement actual balance retrieval from BSC node
	// This is a mock implementation that returns "0"
	// In a real implementation, you would:
	// 1. Connect to a BSC node (Binance API, QuickNode, etc.)
	// 2. For BNB: Use ethclient.BalanceAt() to get the balance
	// 3. For BEP-20: Use the contract's balanceOf function
	// 4. Convert from Wei to BNB or token decimals
	return "0", nil
}

// SendTransaction sends a transaction on the BSC network
func (b *BSCChain) SendTransaction(ctx context.Context, from, to string, amount string, token string, privateKey string) (string, error) {
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

	// Normalize token - empty means BNB
	token = strings.TrimSpace(token)
	if token == "" {
		token = "BNB"
	}

	// Handle token validation
	isERC20 := false
	if strings.ToUpper(token) != "BNB" {
		// Check if it's a valid contract address for BEP-20 token (similar to ERC-20)
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

	// TODO: Implement actual BSC transaction creation and signing
	// This is an enhanced mock implementation with proper validation
	// In a real implementation, you would:
	// 1. Parse amount to Wei (for BNB) or token decimals (for BEP-20)
	// 2. Get current nonce for the from address
	// 3. Estimate gas for the transaction (BSC has different gas mechanics)
	// 4. For BEP-20: Create contract call data for transfer function
	// 5. Create transaction with proper gas price and gas limit for BSC
	// 6. Sign the transaction with the private key
	// 7. Broadcast the transaction to the BSC network
	// 8. Return the actual transaction hash

	// Generate a realistic-looking transaction hash for demo purposes
	var hashInput string
	if isERC20 {
		hashInput = fmt.Sprintf("BSC-BEP20-%s-%s%s%s", token, from, to, amount)
	} else {
		hashInput = fmt.Sprintf("BSC-%s%s%s", from, to, amount)
	}
	hash := crypto.Keccak256Hash([]byte(hashInput))
	return hash.Hex(), nil
}

// EstimateGas estimates gas requirements for a BSC transaction
func (b *BSCChain) EstimateGas(ctx context.Context, from, to string, amount string, token string) (gasLimit uint64, gasPrice string, err error) {
	// Validate addresses
	if !common.IsHexAddress(from) {
		return 0, "", errors.New("invalid from address format")
	}
	if !common.IsHexAddress(to) {
		return 0, "", errors.New("invalid to address format")
	}

	// Normalize token - empty means BNB
	token = strings.TrimSpace(token)
	if token == "" {
		token = "BNB"
	}

	// Basic gas estimation based on transaction type for BSC
	var baseGasLimit uint64
	var baseGasPrice string = "5" // 5 gwei as default for BSC (typically lower than ETH)

	if strings.ToUpper(token) == "BNB" {
		// Simple BNB transfer
		baseGasLimit = 21000
	} else {
		// BEP-20 token transfer requires more gas
		if !common.IsHexAddress(token) {
			return 0, "", fmt.Errorf("invalid token contract address: %s", token)
		}
		baseGasLimit = 65000 // Typical gas for BEP-20 transfer
	}

	// TODO: In a real implementation, you would:
	// 1. Connect to a BSC node
	// 2. Use eth_estimateGas to get actual gas estimate (BSC uses same RPC interface)
	// 3. Get current gas price from the BSC network
	// 4. Apply safety multipliers (e.g., 1.2x for gas limit)
	// 5. Check for BSC network congestion and adjust gas price

	return baseGasLimit, baseGasPrice, nil
}

// ConfirmTransaction checks the confirmation status of a BSC transaction
func (b *BSCChain) ConfirmTransaction(ctx context.Context, txHash string, requiredConfirmations uint64) (*TransactionConfirmation, error) {
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
		requiredConfirmations = 3 // Default for BSC (faster than Ethereum)
	}

	// TODO: Implement actual transaction confirmation checking for BSC
	// This is a mock implementation for development purposes
	// In a real implementation, you would:
	// 1. Connect to a BSC node (BSC RPC endpoints)
	// 2. Get transaction receipt using client.TransactionReceipt(ctx, common.HexToHash(txHash))
	// 3. Get current block number using client.BlockNumber(ctx)
	// 4. Calculate confirmations = currentBlock - txReceipt.BlockNumber
	// 5. Get block details for timestamp using client.BlockByNumber()
	// 6. Return actual transaction details

	// For development purposes, simulate different transaction states
	// This will be replaced with actual blockchain queries
	var status string
	var confirmations uint64
	var blockNumber uint64 = 34567890 // Mock BSC block number (higher than ETH)
	var gasUsed string = "21000"      // Standard BNB transfer gas
	var transactionFee string = "0.000105" // Mock fee (5 gwei * 21000 gas - BSC is cheaper)
	var timestamp time.Time = time.Now().Add(-5 * time.Minute) // Mock timestamp (BSC is faster)

	// Simple logic to simulate transaction states based on hash
	hashBytes := common.HexToHash(txHash).Bytes()
	lastByte := hashBytes[len(hashBytes)-1]
	
	switch {
	case lastByte%12 == 0: // ~8% chance of failed transaction (BSC has better reliability)
		status = "failed"
		confirmations = 0
	case lastByte%4 == 0: // 25% chance of pending transaction
		status = "pending" 
		confirmations = uint64(lastByte) % requiredConfirmations // Random confirmations < required
	default: // ~67% chance of confirmed transaction (BSC confirms faster)
		status = "confirmed"
		confirmations = requiredConfirmations + uint64(lastByte)%5 // Confirmed with extra confirmations
	}

	// Mock different gas usage for different transaction types
	if strings.Contains(strings.ToLower(txHash), "bep20") || lastByte%4 == 1 {
		gasUsed = "52000" // BEP-20 transfer gas
		transactionFee = "0.000260" // Higher fee for token transfers
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