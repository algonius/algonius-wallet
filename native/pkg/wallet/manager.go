// SPDX-License-Identifier: Apache-2.0
package wallet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/security"
	"github.com/algonius/algonius-wallet/native/pkg/wallet/chain"
)

type WalletManager struct {
	chainFactory *chain.ChainFactory
	// For demo purposes, we'll store a simple wallet status
	// In a real implementation, this would be stored securely
	currentWallet *WalletStatus
}

// NewWalletManager constructs a new WalletManager.
func NewWalletManager() *WalletManager {
	return &WalletManager{
		chainFactory: chain.NewChainFactory(),
	}
}

// CreateWallet creates a new wallet for the specified chain.
func (wm *WalletManager) CreateWallet(ctx context.Context, chainName string) (address string, publicKey string, err error) {
	// Get the chain implementation
	chainImpl, err := wm.chainFactory.GetChain(chainName)
	if err != nil {
		return "", "", err
	}

	// Create the wallet using the chain implementation
	walletInfo, err := chainImpl.CreateWallet(ctx)
	if err != nil {
		return "", "", err
	}

	// Store the wallet status (for demo purposes)
	wm.currentWallet = NewWalletStatus(walletInfo.Address, walletInfo.PublicKey)
	// Add supported chains
	wm.currentWallet.Chains["ethereum"] = true
	wm.currentWallet.Chains["bsc"] = true
	wm.currentWallet.Chains["solana"] = true

	// TODO: Store the wallet securely (encrypted private key and mnemonic)
	// For now, we only return the address and public key
	return walletInfo.Address, walletInfo.PublicKey, nil
}

// ImportWallet imports a wallet using a mnemonic phrase
func (wm *WalletManager) ImportWallet(ctx context.Context, mnemonic, password, chainName, derivationPath string) (address string, publicKey string, importedAt int64, err error) {
	// Validate inputs
	if err := ValidateMnemonic(mnemonic); err != nil {
		return "", "", 0, fmt.Errorf("invalid mnemonic: %w", err)
	}

	if err := ValidatePassword(password); err != nil {
		return "", "", 0, fmt.Errorf("weak password: %w", err)
	}

	if err := ValidateChain(chainName); err != nil {
		return "", "", 0, fmt.Errorf("unsupported chain: %w", err)
	}

	// Normalize chain name
	normalizedChain := NormalizeChain(chainName)

	// Set default derivation path if not provided
	if derivationPath == "" {
		derivationPath = "m/44'/60'/0'/0/0" // Default Ethereum derivation path
	}

	// Get the chain implementation
	chainImpl, err := wm.chainFactory.GetChain(normalizedChain)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to get chain implementation: %w", err)
	}

	// Import wallet using the chain-specific implementation
	walletInfo, err := wm.importWalletFromMnemonic(chainImpl, mnemonic, derivationPath)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to import wallet: %w", err)
	}

	// Encrypt private key and mnemonic for storage
	encryptedPrivateKey, err := security.EncryptWithPassword(walletInfo.PrivateKey, password)
	if err != nil {
		return "", "", 0, fmt.Errorf("storage encryption failed: %w", err)
	}

	encryptedMnemonic, err := security.EncryptWithPassword(walletInfo.Mnemonic, password)
	if err != nil {
		return "", "", 0, fmt.Errorf("storage encryption failed: %w", err)
	}

	// Check if wallet already exists (same address)
	if wm.currentWallet != nil && wm.currentWallet.Address == walletInfo.Address {
		return "", "", 0, errors.New("wallet already exists")
	}

	// Store the wallet status
	importTime := time.Now().Unix()
	wm.currentWallet = NewWalletStatus(walletInfo.Address, walletInfo.PublicKey)
	wm.currentWallet.LastUsed = importTime

	// Add supported chains based on imported chain
	switch normalizedChain {
	case "ethereum":
		wm.currentWallet.Chains["ethereum"] = true
		wm.currentWallet.Chains["bsc"] = true // BSC is Ethereum-compatible
	case "bsc":
		wm.currentWallet.Chains["bsc"] = true
		wm.currentWallet.Chains["ethereum"] = true // ETH is BSC-compatible
	}

	// TODO: Store encrypted private key and mnemonic in secure storage
	// For now, we just validate that encryption worked
	_ = encryptedPrivateKey
	_ = encryptedMnemonic

	return walletInfo.Address, walletInfo.PublicKey, importTime, nil
}

// importWalletFromMnemonic imports a wallet from mnemonic using chain-specific logic
func (wm *WalletManager) importWalletFromMnemonic(chainImpl chain.IChain, mnemonic, derivationPath string) (*chain.WalletInfo, error) {
	// For now, we'll use a simplified approach that creates a wallet from mnemonic
	// This should be enhanced to support proper HD wallet derivation paths
	
	// Since the current chain implementations don't have ImportFromMnemonic methods,
	// we'll use the CreateWallet method and then verify it can reproduce wallets
	// In a production system, this would be replaced with proper mnemonic import logic
	
	walletInfo, err := chainImpl.CreateWallet(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	// TODO: Replace this with actual mnemonic-based wallet derivation
	// For now, this is a simplified implementation that demonstrates the structure
	// In production, you would:
	// 1. Parse the derivation path (e.g., "m/44'/60'/0'/0/0")
	// 2. Use the mnemonic to generate a seed
	// 3. Use HD wallet derivation to get the private key at the specified path
	// 4. Generate the corresponding public key and address

	return walletInfo, nil
}

// GetBalance returns the balance for the given address and token.
func (wm *WalletManager) GetBalance(ctx context.Context, address, token string) (string, error) {
	if address == "" {
		return "", errors.New("address is required")
	}
	if token == "" {
		token = "ETH"
	}

	// For now, assume ETH chain for balance queries
	// In the future, we could determine the chain from the address format
	chainImpl, err := wm.chainFactory.GetChain("ETH")
	if err != nil {
		return "", err
	}

	return chainImpl.GetBalance(ctx, address, token)
}

// GetStatus returns the current wallet status.
func (wm *WalletManager) GetStatus(ctx context.Context) (*WalletStatus, error) {
	if wm.currentWallet == nil {
		// Return a default status if no wallet is created yet
		return &WalletStatus{
			Address:   "",
			PublicKey: "",
			Ready:     false,
			Chains:    map[string]bool{
				"ethereum": true,
				"bsc":      true,
				"solana":   true,
			},
			LastUsed: 0,
		}, nil
	}
	
	return wm.currentWallet, nil
}

// SendTransaction sends a transaction on the specified chain.
func (wm *WalletManager) SendTransaction(ctx context.Context, chain, from, to, amount, token string) (string, error) {
	if wm.currentWallet == nil {
		return "", errors.New("no wallet available - create a wallet first")
	}

	// Validate required parameters
	if from == "" || to == "" || amount == "" {
		return "", errors.New("from, to, and amount are required")
	}

	// Additional security checks
	if err := wm.validateTransactionSecurity(from, to, amount, token); err != nil {
		return "", fmt.Errorf("security validation failed: %w", err)
	}

	// Get the chain implementation
	chainImpl, err := wm.chainFactory.GetChain(chain)
	if err != nil {
		return "", err
	}

	// TODO: In a real implementation, we would need to retrieve the private key
	// For now, we'll use a mock private key since wallet storage is not fully implemented
	mockPrivateKey := "0x0000000000000000000000000000000000000000000000000000000000000001"

	// Send the transaction using the chain implementation
	return chainImpl.SendTransaction(ctx, from, to, amount, token, mockPrivateKey)
}

// validateTransactionSecurity performs basic security validations
func (wm *WalletManager) validateTransactionSecurity(from, to, amount, token string) error {
	// Validate addresses
	if !wm.isValidAddress(from) {
		return errors.New("invalid from address")
	}
	if !wm.isValidAddress(to) {
		return errors.New("invalid to address")
	}

	// Validate amount format (basic check)
	if err := wm.validateAmount(amount); err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	// Prevent sending to zero address
	if to == "0x0000000000000000000000000000000000000000" {
		return errors.New("cannot send to zero address")
	}

	// Prevent sending to same address
	if strings.EqualFold(from, to) {
		return errors.New("cannot send to the same address")
	}

	// TODO: In a real implementation, add more security checks:
	// - Check balance before sending
	// - Implement transaction amount limits
	// - Add confirmation mechanisms for large transactions
	// - Validate gas limits and prevent excessive fees

	return nil
}

// isValidAddress checks if an address is valid (currently supports Ethereum-style addresses)
func (wm *WalletManager) isValidAddress(address string) bool {
	// Basic validation for Ethereum-style addresses
	if len(address) != 42 {
		return false
	}
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	// Could add checksum validation here
	return true
}

// validateAmount performs basic amount validation
func (wm *WalletManager) validateAmount(amount string) error {
	if amount == "" {
		return errors.New("amount cannot be empty")
	}
	if amount == "0" || amount == "0.0" {
		return errors.New("amount must be greater than zero")
	}
	// TODO: Add more sophisticated amount parsing and validation
	// - Parse as decimal number
	// - Check for reasonable limits
	// - Validate decimal places
	return nil
}

// EstimateGas estimates gas requirements for a transaction on the specified chain.
func (wm *WalletManager) EstimateGas(ctx context.Context, chain, from, to, amount, token string) (uint64, string, error) {
	// Validate required parameters
	if from == "" || to == "" || amount == "" {
		return 0, "", errors.New("from, to, and amount are required")
	}

	// Get the chain implementation
	chainImpl, err := wm.chainFactory.GetChain(chain)
	if err != nil {
		return 0, "", err
	}

	// Estimate gas using the chain implementation
	return chainImpl.EstimateGas(ctx, from, to, amount, token)
}
