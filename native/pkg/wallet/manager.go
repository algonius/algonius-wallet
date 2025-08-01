// SPDX-License-Identifier: Apache-2.0
package wallet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/config"
	"github.com/algonius/algonius-wallet/native/pkg/dex"
	"github.com/algonius/algonius-wallet/native/pkg/security"
	"github.com/algonius/algonius-wallet/native/pkg/wallet/chain"
	"go.uber.org/zap"
)

// EncryptedWalletData represents encrypted wallet storage format
type EncryptedWalletData struct {
	Address          string                 `json:"address"`
	PublicKey        string                 `json:"public_key"`
	EncryptedPrivateKey *security.EncryptedData `json:"encrypted_private_key"`
	EncryptedMnemonic    *security.EncryptedData `json:"encrypted_mnemonic"`
	Chains           map[string]bool        `json:"chains"`
	CreatedAt        int64                  `json:"created_at"`
	LastUsed         int64                  `json:"last_used"`
}

// DecryptedWalletData represents decrypted wallet data in memory
type DecryptedWalletData struct {
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Mnemonic   string `json:"mnemonic"`
}

type WalletManager struct {
	chainFactory *chain.ChainFactory
	// Storage configuration
	walletDir    string
	// Current wallet state (only loaded when user enters password)
	currentWallet *WalletStatus
	currentWalletData *DecryptedWalletData
	isUnlocked   bool
	// Audit logger for security events
	auditLogger *AuditLogger
	// Mock storage for pending transactions
	pendingTxs []*PendingTransaction
	// Logger for debugging and monitoring
	logger *zap.Logger
}

// NewWalletManager constructs a new WalletManager.
func NewWalletManager() *WalletManager {
	// Get wallet directory from environment or default
	walletHomeDir := getWalletHomeDir()
	walletDir := filepath.Join(walletHomeDir, "wallets")
	
	// Create wallet directory if it doesn't exist
	os.MkdirAll(walletDir, 0700)
	
	// Use no-op logger if none provided
	logger := zap.NewNop()
	
	return &WalletManager{
		chainFactory: chain.NewChainFactory(),
		walletDir:    walletDir,
		auditLogger:  NewAuditLogger(),
		pendingTxs:   make([]*PendingTransaction, 0),
		isUnlocked:   false,
		logger:       logger,
	}
}

// NewWalletManagerWithConfig constructs a new WalletManager with configuration.
func NewWalletManagerWithConfig(config *config.Config, dexAggregator dex.IDEXAggregator, logger *zap.Logger) *WalletManager {
	// Get wallet directory from config or environment
	logger.Info("NewWalletManagerWithConfig starting", 
		zap.String("config_data_dir", config.Wallet.DataDir))
	
	// Ensure the DataDir path is properly expanded (handle ~ if present)
	dataDir := config.Wallet.DataDir
	if dataDir != "" && dataDir[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logger.Error("Failed to get home directory", zap.Error(err))
			dataDir = "." // fallback to current directory
		} else {
			dataDir = filepath.Join(homeDir, dataDir[1:])
		}
	}
	
	walletDir := filepath.Join(dataDir, "wallets")
	logger.Info("NewWalletManagerWithConfig wallet directory constructed", 
		zap.String("wallet_dir", walletDir),
		zap.String("expanded_data_dir", dataDir))
	
	// Create wallet directory if it doesn't exist
	os.MkdirAll(walletDir, 0700)
	
	// Create chain factory with configuration
	var chainFactory *chain.ChainFactory
	if dexAggregator != nil && logger != nil {
		chainFactory = chain.NewChainFactoryWithDEX(dexAggregator, logger, config)
	} else {
		chainFactory = chain.NewChainFactory()
	}
	
	return &WalletManager{
		chainFactory: chainFactory,
		walletDir:    walletDir,
		auditLogger:  NewAuditLogger(),
		pendingTxs:   make([]*PendingTransaction, 0),
		isUnlocked:   false,
		logger:       logger,
	}
}

// getWalletHomeDir returns the wallet home directory, respecting environment override
func getWalletHomeDir() string {
	// Check environment variable first
	if homeDir := os.Getenv("ALGONIUS_WALLET_HOME"); homeDir != "" {
		return homeDir
	}
	
	// Default path
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(userHome, ".algonius-wallet")
}

// CreateWallet creates a new wallet for the specified chain.
func (wm *WalletManager) CreateWallet(ctx context.Context, chainName, password string) (address string, publicKey string, mnemonic string, err error) {
	wm.logger.Info("WalletManager.CreateWallet started", 
		zap.String("chain", chainName),
		zap.Int("password_length", len(password)))
	
	// Validate password
	if err := ValidatePassword(password); err != nil {
		wm.logger.Error("CreateWallet password validation failed", 
			zap.Error(err),
			zap.String("chain", chainName))
		return "", "", "", fmt.Errorf("weak password: %w", err)
	}
	wm.logger.Info("CreateWallet password validation passed")

	// Validate and normalize chain name
	if err := ValidateChain(chainName); err != nil {
		wm.logger.Error("CreateWallet chain validation failed", 
			zap.Error(err),
			zap.String("chain", chainName))
		return "", "", "", err
	}
	normalizedChain := NormalizeChain(chainName)
	wm.logger.Info("CreateWallet chain validation passed", 
		zap.String("original_chain", chainName),
		zap.String("normalized_chain", normalizedChain))

	// Get the chain implementation
	chainImpl, err := wm.chainFactory.GetChain(normalizedChain)
	if err != nil {
		wm.logger.Error("CreateWallet failed to get chain implementation", 
			zap.Error(err),
			zap.String("normalized_chain", normalizedChain))
		return "", "", "", err
	}
	wm.logger.Info("CreateWallet got chain implementation", 
		zap.String("chain_name", chainImpl.GetChainName()))

	// Create the wallet using the chain implementation
	wm.logger.Info("CreateWallet calling chain.CreateWallet")
	walletInfo, err := chainImpl.CreateWallet(ctx)
	if err != nil {
		wm.logger.Error("CreateWallet chain.CreateWallet failed", 
			zap.Error(err),
			zap.String("chain", normalizedChain))
		return "", "", "", err
	}
	wm.logger.Info("CreateWallet chain.CreateWallet succeeded", 
		zap.String("address", walletInfo.Address),
		zap.String("public_key", walletInfo.PublicKey[:20]+"..."),
		zap.Int("mnemonic_word_count", len(strings.Fields(walletInfo.Mnemonic))))

	// Encrypt private key and mnemonic for storage
	wm.logger.Info("CreateWallet starting encryption")
	encryptedPrivateKey, err := security.EncryptWithPassword(walletInfo.PrivateKey, password)
	if err != nil {
		wm.logger.Error("CreateWallet private key encryption failed", 
			zap.Error(err))
		return "", "", "", fmt.Errorf("storage encryption failed: %w", err)
	}

	encryptedMnemonic, err := security.EncryptWithPassword(walletInfo.Mnemonic, password)
	if err != nil {
		wm.logger.Error("CreateWallet mnemonic encryption failed", 
			zap.Error(err))
		return "", "", "", fmt.Errorf("storage encryption failed: %w", err)
	}
	wm.logger.Info("CreateWallet encryption completed")

	// Store the wallet status
	createdTime := time.Now().Unix()
	wm.currentWallet = NewWalletStatus(walletInfo.Address, walletInfo.PublicKey)
	wm.currentWallet.LastUsed = createdTime

	// Add supported chains based on created chain
	switch normalizedChain {
	case "ethereum":
		wm.currentWallet.Chains["ethereum"] = true
		wm.currentWallet.Chains["bsc"] = true // BSC is Ethereum-compatible
	case "bsc":
		wm.currentWallet.Chains["bsc"] = true
		wm.currentWallet.Chains["ethereum"] = true // ETH is BSC-compatible
	}

	// Create encrypted wallet data structure
	encryptedWallet := &EncryptedWalletData{
		Address:             walletInfo.Address,
		PublicKey:           walletInfo.PublicKey,
		EncryptedPrivateKey: encryptedPrivateKey,
		EncryptedMnemonic:   encryptedMnemonic,
		Chains:              make(map[string]bool),
		CreatedAt:           createdTime,
		LastUsed:            createdTime,
	}
	
	// Add supported chains to encrypted wallet data
	switch normalizedChain {
	case "ethereum":
		encryptedWallet.Chains["ethereum"] = true
		encryptedWallet.Chains["bsc"] = true
	case "bsc":
		encryptedWallet.Chains["bsc"] = true
		encryptedWallet.Chains["ethereum"] = true
	}
	
	// Save encrypted wallet to disk
	wm.logger.Info("CreateWallet saving wallet to disk")
	err = wm.saveWalletToDisk(encryptedWallet)
	if err != nil {
		wm.logger.Error("CreateWallet failed to save wallet to disk", 
			zap.Error(err))
		return "", "", "", fmt.Errorf("failed to save wallet: %w", err)
	}
	wm.logger.Info("CreateWallet wallet saved to disk successfully")
	
	// Load decrypted data into memory for immediate use
	wm.logger.Info("CreateWallet loading wallet into memory and unlocking")
	wm.currentWalletData = &DecryptedWalletData{
		Address:    walletInfo.Address,
		PublicKey:  walletInfo.PublicKey,
		PrivateKey: walletInfo.PrivateKey,
		Mnemonic:   walletInfo.Mnemonic,
	}
	wm.isUnlocked = true

	wm.logger.Info("CreateWallet completed successfully", 
		zap.String("address", walletInfo.Address),
		zap.Int("mnemonic_word_count", len(strings.Fields(walletInfo.Mnemonic))))

	return walletInfo.Address, walletInfo.PublicKey, walletInfo.Mnemonic, nil
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
	walletInfo, err := wm.importWalletFromMnemonic(ctx, chainImpl, mnemonic, derivationPath)
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

	// Create encrypted wallet data structure
	encryptedWallet := &EncryptedWalletData{
		Address:             walletInfo.Address,
		PublicKey:           walletInfo.PublicKey,
		EncryptedPrivateKey: encryptedPrivateKey,
		EncryptedMnemonic:   encryptedMnemonic,
		Chains:              make(map[string]bool),
		CreatedAt:           importTime,
		LastUsed:            importTime,
	}
	
	// Add supported chains based on imported chain
	switch normalizedChain {
	case "ethereum":
		encryptedWallet.Chains["ethereum"] = true
		encryptedWallet.Chains["bsc"] = true // BSC is Ethereum-compatible
	case "bsc":
		encryptedWallet.Chains["bsc"] = true
		encryptedWallet.Chains["ethereum"] = true // ETH is BSC-compatible
	}
	
	// Save encrypted wallet to disk
	err = wm.saveWalletToDisk(encryptedWallet)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to save wallet: %w", err)
	}
	
	// Load decrypted data into memory for immediate use
	wm.currentWalletData = &DecryptedWalletData{
		Address:    walletInfo.Address,
		PublicKey:  walletInfo.PublicKey,
		PrivateKey: walletInfo.PrivateKey,
		Mnemonic:   walletInfo.Mnemonic,
	}
	wm.isUnlocked = true

	return walletInfo.Address, walletInfo.PublicKey, importTime, nil
}

// importWalletFromMnemonic imports a wallet from mnemonic using chain-specific logic
func (wm *WalletManager) importWalletFromMnemonic(ctx context.Context, chainImpl chain.IChain, mnemonic, derivationPath string) (*chain.WalletInfo, error) {
	// Use the chain-specific ImportFromMnemonic method
	walletInfo, err := chainImpl.ImportFromMnemonic(ctx, mnemonic, derivationPath)
	if err != nil {
		return nil, fmt.Errorf("failed to import wallet from mnemonic: %w", err)
	}

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

	// Determine chain based on token identifier
	// This is a simple approach - in a production system, you might want a more sophisticated
	// method that can determine the chain from the address format or other metadata
	var chainName string
	tokenUpper := strings.ToUpper(token)
	
	switch {
	case tokenUpper == "ETH" || tokenUpper == "ETHER":
		chainName = "ETH"
	case tokenUpper == "BNB" || tokenUpper == "BINANCE":
		chainName = "BSC"
	case tokenUpper == "SOL" || tokenUpper == "SOLANA":
		chainName = "SOL"
	default:
		// For contract addresses, we'll default to ETH for now
		// A more sophisticated implementation might analyze the address format
		// or require an explicit chain parameter
		chainName = "ETH"
	}

	// Get the chain implementation
	chainImpl, err := wm.chainFactory.GetChain(chainName)
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

// GetPendingTransactions retrieves pending transactions with optional filtering and pagination
func (wm *WalletManager) GetPendingTransactions(ctx context.Context, chain, address, transactionType string, limit, offset int) ([]*PendingTransaction, error) {
	// For now, we'll return mock pending transactions for development purposes
	// In a real implementation, this would:
	// 1. Query the blockchain network for pending transactions
	// 2. Filter by owned wallet addresses
	// 3. Apply the specified filters (chain, address, type)
	// 4. Return paginated results
	
	// Validate parameters
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	
	// Generate mock pending transactions for development
	mockTxs := wm.generateMockPendingTransactions(chain, address, transactionType)
	
	// Apply pagination
	start := offset
	if start >= len(mockTxs) {
		return []*PendingTransaction{}, nil
	}
	
	end := start + limit
	if end > len(mockTxs) {
		end = len(mockTxs)
	}
	
	return mockTxs[start:end], nil
}

// generateMockPendingTransactions creates mock pending transactions for development
func (wm *WalletManager) generateMockPendingTransactions(chain, address, transactionType string) []*PendingTransaction {
	baseTime := time.Now()
	
	// Create a variety of mock transactions
	mockTxs := []*PendingTransaction{
		{
			Hash:                      "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
			Chain:                     "ethereum",
			From:                      "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			To:                        "0x8ba1f109551bD432803012645Hac136c22C4F9B",
			Amount:                    "0.5",
			Token:                     "ETH",
			Type:                      "transfer",
			Status:                    "pending",
			Confirmations:             2,
			RequiredConfirmations:     6,
			BlockNumber:               18500123,
			Nonce:                     42,
			GasFee:                    "0.0021",
			Priority:                  "medium",
			EstimatedConfirmationTime: "2-3 minutes",
			SubmittedAt:               baseTime.Add(-5 * time.Minute),
			LastChecked:               baseTime.Add(-30 * time.Second),
		},
		{
			Hash:                      "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			Chain:                     "bsc",
			From:                      "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			To:                        "0x0000000000000000000000000000000000000000",
			Amount:                    "1000",
			Token:                     "0x55d398326f99059fF775485246999027B3197955", // USDT contract
			Type:                      "swap",
			Status:                    "pending",
			Confirmations:             0,
			RequiredConfirmations:     3,
			BlockNumber:               0,
			Nonce:                     43,
			GasFee:                    "0.0008",
			Priority:                  "high",
			EstimatedConfirmationTime: "30-60 seconds",
			SubmittedAt:               baseTime.Add(-2 * time.Minute),
			LastChecked:               baseTime.Add(-15 * time.Second),
		},
		{
			Hash:                      "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
			Chain:                     "ethereum",
			From:                      "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			To:                        "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984",
			Amount:                    "50",
			Token:                     "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984", // UNI contract
			Type:                      "contract",
			Status:                    "pending",
			Confirmations:             4,
			RequiredConfirmations:     6,
			BlockNumber:               18500125,
			Nonce:                     44,
			GasFee:                    "0.0042",
			Priority:                  "low",
			EstimatedConfirmationTime: "5-10 minutes",
			SubmittedAt:               baseTime.Add(-8 * time.Minute),
			LastChecked:               baseTime.Add(-45 * time.Second),
		},
	}
	
	// Apply filters
	var filteredTxs []*PendingTransaction
	
	for _, tx := range mockTxs {
		// Filter by chain
		if chain != "" && strings.ToLower(tx.Chain) != strings.ToLower(chain) {
			continue
		}
		
		// Filter by address (from or to)
		if address != "" && 
			!strings.EqualFold(tx.From, address) && 
			!strings.EqualFold(tx.To, address) {
			continue
		}
		
		// Filter by transaction type
		if transactionType != "" && strings.ToLower(tx.Type) != strings.ToLower(transactionType) {
			continue
		}
		
		filteredTxs = append(filteredTxs, tx)
	}
	
	return filteredTxs
}

// RejectTransactions rejects multiple pending transactions with specified reasons
func (wm *WalletManager) RejectTransactions(ctx context.Context, transactionIds []string, reason, details string, notifyUser, auditLog bool) ([]TransactionRejectionResult, error) {
	results := make([]TransactionRejectionResult, 0, len(transactionIds))
	
	for _, txHash := range transactionIds {
		result := TransactionRejectionResult{
			TransactionHash: txHash,
			Success:         false,
		}

		// Find the transaction in mock data
		var foundTx *PendingTransaction
		for _, mockTx := range wm.generateMockPendingTransactions("", "", "") {
			if mockTx.Hash == txHash {
				foundTx = mockTx
				break
			}
		}

		if foundTx == nil {
			result.ErrorMessage = "transaction not found"
			results = append(results, result)
			continue
		}

		// Check if transaction is already rejected or completed
		if foundTx.Status == "rejected" {
			result.ErrorMessage = "transaction already rejected"
			results = append(results, result)
			continue
		}

		if foundTx.Status == "confirmed" {
			result.ErrorMessage = "cannot reject confirmed transaction"
			results = append(results, result)
			continue
		}

		// Validate transaction ownership (basic check)
		if wm.currentWallet != nil && foundTx.From != wm.currentWallet.Address {
			result.ErrorMessage = "unauthorized: transaction does not belong to current wallet"
			results = append(results, result)
			continue
		}

		// Perform the rejection
		rejectionTime := time.Now()
		
		// Update transaction status
		foundTx.Status = "rejected"
		foundTx.RejectedAt = &rejectionTime
		foundTx.RejectionReason = reason
		foundTx.RejectionDetails = details

		// Log to audit trail if requested
		var auditLogId string
		if auditLog {
			logId, err := wm.auditLogger.LogTransactionRejection(txHash, reason, details, foundTx.From)
			if err != nil {
				result.ErrorMessage = fmt.Sprintf("audit logging failed: %v", err)
				results = append(results, result)
				continue
			}
			auditLogId = logId
			foundTx.RejectionAuditLogId = auditLogId
		}

		// Send user notification if requested
		if notifyUser {
			// In a real implementation, this would send actual notifications
			// For now, we just simulate the action
			_ = wm.sendRejectionNotification(foundTx, reason, details)
		}

		// Mark as successful
		result.Success = true
		result.RejectedAt = rejectionTime
		result.AuditLogId = auditLogId

		results = append(results, result)
	}

	return results, nil
}

// sendRejectionNotification simulates sending a notification to the user
func (wm *WalletManager) sendRejectionNotification(tx *PendingTransaction, reason, details string) error {
	// In a real implementation, this would:
	// - Send email notification
	// - Send push notification
	// - Log notification to user activity feed
	// - Possibly send SMS for high-value transactions
	
	// For now, we just simulate success
	return nil
}

// GetTransactionHistory retrieves historical transactions for the specified address with optional filtering
func (wm *WalletManager) GetTransactionHistory(ctx context.Context, address string, fromBlock, toBlock *uint64, limit, offset int) ([]*HistoricalTransaction, error) {
	// Validate required parameters
	if address == "" {
		return nil, errors.New("address is required")
	}
	
	// Validate pagination parameters
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	
	// For now, we'll return mock historical transactions for development purposes
	// In a real implementation, this would:
	// 1. Query blockchain RPC endpoints for transaction history
	// 2. Parse transaction logs for ERC-20/SPL token transfers
	// 3. Decode contract interactions
	// 4. Apply block range filters
	// 5. Return paginated results
	
	// Generate mock historical transactions
	mockTxs := wm.generateMockHistoricalTransactions(address, fromBlock, toBlock)
	
	// Apply pagination
	start := offset
	if start >= len(mockTxs) {
		return []*HistoricalTransaction{}, nil
	}
	
	end := start + limit
	if end > len(mockTxs) {
		end = len(mockTxs)
	}
	
	return mockTxs[start:end], nil
}

// generateMockHistoricalTransactions creates mock historical transactions for development
func (wm *WalletManager) generateMockHistoricalTransactions(address string, fromBlock, toBlock *uint64) []*HistoricalTransaction {
	baseTime := time.Now()
	
	// Create a variety of mock historical transactions
	mockTxs := []*HistoricalTransaction{
		{
			Hash:              "0x123abc456def789ghi012jkl345mno678pqr901stu234vwx567yza890bcd123efg",
			Chain:             "ethereum",
			BlockNumber:       18500100,
			BlockHash:         "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			TransactionIndex:  42,
			From:              "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			To:                "0x8ba1f109551bD432803012645Hac136c22C4F9B",
			Value:             "1.5",
			Token:             "ETH",
			TokenSymbol:       "ETH",
			Type:              "transfer",
			Status:            "confirmed",
			GasUsed:           "21000",
			GasPrice:          "20000000000",
			TransactionFee:    "0.00042",
			Timestamp:         baseTime.Add(-2 * time.Hour),
			Confirmations:     50,
		},
		{
			Hash:              "0x789def012abc345ghi678jkl901mno234pqr567stu890vwx123yza456bcd789efg",
			Chain:             "ethereum",
			BlockNumber:       18500095,
			BlockHash:         "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
			TransactionIndex:  15,
			From:              "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			To:                "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984",
			Value:             "0",
			Token:             "0xA0b86a33E6441C8C606A57B0e25A3F8A7ad0a93D", // UNI token
			TokenSymbol:       "UNI",
			Type:              "contract_call",
			Status:            "confirmed",
			GasUsed:           "65432",
			GasPrice:          "18000000000",
			TransactionFee:    "0.001177776",
			Timestamp:         baseTime.Add(-4 * time.Hour),
			Confirmations:     55,
			ContractAddress:   "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984",
			MethodName:        "transfer",
			TokenTransfers: []TokenTransfer{
				{
					From:          "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
					To:            "0x8ba1f109551bD432803012645Hac136c22C4F9B",
					Value:         "100",
					TokenAddress:  "0xA0b86a33E6441C8C606A57B0e25A3F8A7ad0a93D",
					TokenSymbol:   "UNI",
					TokenDecimals: 18,
				},
			},
		},
		{
			Hash:              "0x456ghi789jkl012mno345pqr678stu901vwx234yza567bcd890efg123abc456def",
			Chain:             "bsc",
			BlockNumber:       32150200,
			BlockHash:         "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			TransactionIndex:  8,
			From:              "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			To:                "0x10ed43c718714eb63d5aa57b78b54704e256024e", // PancakeSwap Router
			Value:             "0.1",
			Token:             "BNB",
			TokenSymbol:       "BNB",
			Type:              "swap",
			Status:            "confirmed",
			GasUsed:           "180000",
			GasPrice:          "5000000000",
			TransactionFee:    "0.0009",
			Timestamp:         baseTime.Add(-6 * time.Hour),
			Confirmations:     120,
			ContractAddress:   "0x10ed43c718714eb63d5aa57b78b54704e256024e",
			MethodName:        "swapExactETHForTokens",
			TokenTransfers: []TokenTransfer{
				{
					From:          "0x10ed43c718714eb63d5aa57b78b54704e256024e",
					To:            "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
					Value:         "1500.25",
					TokenAddress:  "0x55d398326f99059fF775485246999027B3197955",
					TokenSymbol:   "USDT",
					TokenDecimals: 18,
				},
			},
		},
		{
			Hash:              "0xabc123def456ghi789jkl012mno345pqr678stu901vwx234yza567bcd890efg123",
			Chain:             "ethereum",
			BlockNumber:       18500085,
			BlockHash:         "0x9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba",
			TransactionIndex:  3,
			From:              "0x8ba1f109551bD432803012645Hac136c22C4F9B",
			To:                "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			Value:             "2.0",
			Token:             "ETH",
			TokenSymbol:       "ETH",
			Type:              "transfer",
			Status:            "confirmed",
			GasUsed:           "21000",
			GasPrice:          "22000000000",
			TransactionFee:    "0.000462",
			Timestamp:         baseTime.Add(-8 * time.Hour),
			Confirmations:     65,
		},
		{
			Hash:              "0xdef789ghi012jkl345mno678pqr901stu234vwx567yza890bcd123efg456abc789",
			Chain:             "ethereum",
			BlockNumber:       18500080,
			BlockHash:         "0x5555555555555555555555555555555555555555555555555555555555555555",
			TransactionIndex:  27,
			From:              "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			To:                "0x0000000000000000000000000000000000000000",
			Value:             "0",
			Token:             "0x6B175474E89094C44Da98b954EedeAC495271d0F", // DAI
			TokenSymbol:       "DAI",
			Type:              "failed",
			Status:            "failed",
			GasUsed:           "45000",
			GasPrice:          "25000000000",
			TransactionFee:    "0.001125",
			Timestamp:         baseTime.Add(-12 * time.Hour),
			Confirmations:     70,
		},
	}
	
	// Apply block range filters
	var filteredTxs []*HistoricalTransaction
	for _, tx := range mockTxs {
		// Filter by from_block
		if fromBlock != nil && tx.BlockNumber < *fromBlock {
			continue
		}
		
		// Filter by to_block
		if toBlock != nil && tx.BlockNumber > *toBlock {
			continue
		}
		
		// Filter by address (should be either from or to)
		if !strings.EqualFold(tx.From, address) && !strings.EqualFold(tx.To, address) {
			continue
		}
		
		filteredTxs = append(filteredTxs, tx)
	}
	
	return filteredTxs
}

// GetAccounts returns available wallet addresses
func (wm *WalletManager) GetAccounts(ctx context.Context) ([]string, error) {
	// In a real implementation, this would return all available wallet addresses
	// For now, return the current wallet address if available
	if wm.currentWallet == nil || wm.currentWallet.Address == "" {
		return []string{}, nil
	}
	
	return []string{wm.currentWallet.Address}, nil
}

// AddPendingTransaction adds a new pending transaction to the queue
func (wm *WalletManager) AddPendingTransaction(ctx context.Context, tx *PendingTransaction) error {
	if tx == nil {
		return errors.New("transaction cannot be nil")
	}
	
	// Basic validation
	if tx.Hash == "" {
		return errors.New("transaction hash is required")
	}
	if tx.From == "" {
		return errors.New("from address is required")
	}
	if tx.To == "" {
		return errors.New("to address is required")
	}
	
	// Check if transaction already exists
	for _, existing := range wm.pendingTxs {
		if existing.Hash == tx.Hash {
			return fmt.Errorf("transaction with hash %s already exists", tx.Hash)
		}
	}
	
	// Add to pending transactions
	wm.pendingTxs = append(wm.pendingTxs, tx)
	
	return nil
}

// Wallet Storage Methods

// saveWalletToDisk saves encrypted wallet data to disk
func (wm *WalletManager) saveWalletToDisk(walletData *EncryptedWalletData) error {
	walletFile := filepath.Join(wm.walletDir, "wallet.json")
	
	wm.logger.Info("saveWalletToDisk starting", 
		zap.String("wallet_file", walletFile),
		zap.String("wallet_dir", wm.walletDir))
	
	// Ensure wallet directory exists
	if err := os.MkdirAll(wm.walletDir, 0700); err != nil {
		wm.logger.Error("saveWalletToDisk failed to create wallet directory", 
			zap.Error(err),
			zap.String("wallet_dir", wm.walletDir))
		return fmt.Errorf("failed to create wallet directory: %w", err)
	}
	wm.logger.Info("saveWalletToDisk wallet directory exists/created")
	
	// Convert to JSON
	jsonData, err := json.MarshalIndent(walletData, "", "  ")
	if err != nil {
		wm.logger.Error("saveWalletToDisk failed to marshal wallet data", 
			zap.Error(err))
		return fmt.Errorf("failed to marshal wallet data: %w", err)
	}
	wm.logger.Info("saveWalletToDisk wallet data marshaled", 
		zap.Int("json_size", len(jsonData)))
	
	// Write to file with restricted permissions
	err = os.WriteFile(walletFile, jsonData, 0600)
	if err != nil {
		wm.logger.Error("saveWalletToDisk failed to write wallet file", 
			zap.Error(err),
			zap.String("wallet_file", walletFile))
		return fmt.Errorf("failed to write wallet file: %w", err)
	}
	
	wm.logger.Info("saveWalletToDisk completed successfully", 
		zap.String("wallet_file", walletFile))
	return nil
}

// loadWalletFromDisk loads encrypted wallet data from disk
func (wm *WalletManager) loadWalletFromDisk() (*EncryptedWalletData, error) {
	walletFile := filepath.Join(wm.walletDir, "wallet.json")
	
	// Check if wallet file exists
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return nil, errors.New("no wallet found")
	}
	
	// Read wallet file
	jsonData, err := os.ReadFile(walletFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}
	
	// Parse JSON
	var walletData EncryptedWalletData
	err = json.Unmarshal(jsonData, &walletData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse wallet file: %w", err)
	}
	
	return &walletData, nil
}

// UnlockWallet decrypts and loads wallet data into memory with password
func (wm *WalletManager) UnlockWallet(password string) error {
	// Load encrypted wallet data from disk
	encryptedWallet, err := wm.loadWalletFromDisk()
	if err != nil {
		return fmt.Errorf("failed to load wallet: %w", err)
	}
	
	// Decrypt private key
	privateKey, err := security.DecryptWithPassword(encryptedWallet.EncryptedPrivateKey, password)
	if err != nil {
		return fmt.Errorf("incorrect password or corrupted wallet: %w", err)
	}
	
	// Decrypt mnemonic
	mnemonic, err := security.DecryptWithPassword(encryptedWallet.EncryptedMnemonic, password)
	if err != nil {
		return fmt.Errorf("incorrect password or corrupted wallet: %w", err)
	}
	
	// Load decrypted data into memory
	wm.currentWalletData = &DecryptedWalletData{
		Address:    encryptedWallet.Address,
		PublicKey:  encryptedWallet.PublicKey,
		PrivateKey: privateKey,
		Mnemonic:   mnemonic,
	}
	
	// Load wallet status
	wm.currentWallet = &WalletStatus{
		Address:   encryptedWallet.Address,
		PublicKey: encryptedWallet.PublicKey,
		Chains:    encryptedWallet.Chains,
		LastUsed:  encryptedWallet.LastUsed,
	}
	
	wm.isUnlocked = true
	
	return nil
}

// LockWallet clears sensitive data from memory
func (wm *WalletManager) LockWallet() {
	if wm.currentWalletData != nil {
		// Clear sensitive data
		wm.currentWalletData.PrivateKey = ""
		wm.currentWalletData.Mnemonic = ""
		wm.currentWalletData = nil
	}
	wm.isUnlocked = false
}

// IsUnlocked returns whether the wallet is currently unlocked
func (wm *WalletManager) IsUnlocked() bool {
	return wm.isUnlocked && wm.currentWalletData != nil
}

// HasWallet returns whether a wallet file exists on disk
func (wm *WalletManager) HasWallet() bool {
	walletFile := filepath.Join(wm.walletDir, "wallet.json")
	_, err := os.Stat(walletFile)
	return err == nil
}

// GetCurrentWallet returns the current wallet status
func (wm *WalletManager) GetCurrentWallet() *WalletStatus {
	return wm.currentWallet
}
