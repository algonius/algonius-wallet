// SPDX-License-Identifier: Apache-2.0
package chain

import (
	"context"
	"time"
)

// WalletInfo contains the generated wallet information
type WalletInfo struct {
	Address    string
	PublicKey  string
	PrivateKey string // Should be encrypted before storage
	Mnemonic   string // Should be encrypted before storage
}

// TransactionConfirmation contains transaction confirmation details
type TransactionConfirmation struct {
	Status               string    `json:"status"`                 // "pending", "confirmed", "failed"
	Confirmations        uint64    `json:"confirmations"`          // Current number of confirmations
	RequiredConfirmations uint64   `json:"required_confirmations"` // Required confirmations for finality
	BlockNumber          uint64    `json:"block_number"`           // Block number containing the transaction
	GasUsed              string    `json:"gas_used"`               // Gas units used
	TransactionFee       string    `json:"transaction_fee"`        // Transaction fee in native currency
	Timestamp            time.Time `json:"timestamp"`              // Block timestamp
	TxHash               string    `json:"tx_hash"`                // Transaction hash
}

// IChain defines the interface for blockchain-specific operations
type IChain interface {
	// CreateWallet generates a new wallet for the chain
	CreateWallet(ctx context.Context) (*WalletInfo, error)

	// GetBalance retrieves the balance for an address
	GetBalance(ctx context.Context, address string, token string) (string, error)

	// SendTransaction sends a transaction on the chain
	SendTransaction(ctx context.Context, from, to string, amount string, token string, privateKey string) (string, error)

	// EstimateGas estimates gas for a transaction
	EstimateGas(ctx context.Context, from, to string, amount string, token string) (gasLimit uint64, gasPrice string, err error)

	// ConfirmTransaction checks the confirmation status of a transaction
	ConfirmTransaction(ctx context.Context, txHash string, requiredConfirmations uint64) (*TransactionConfirmation, error)

	// GetChainName returns the name of the chain
	GetChainName() string
}
