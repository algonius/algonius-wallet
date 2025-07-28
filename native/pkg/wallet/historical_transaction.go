// SPDX-License-Identifier: Apache-2.0
package wallet

import (
	"time"
)

// HistoricalTransaction represents a confirmed transaction from blockchain history
type HistoricalTransaction struct {
	Hash              string    `json:"hash"`
	Chain             string    `json:"chain"`
	BlockNumber       uint64    `json:"block_number"`
	BlockHash         string    `json:"block_hash,omitempty"`
	TransactionIndex  uint64    `json:"transaction_index,omitempty"`
	From              string    `json:"from"`
	To                string    `json:"to"`
	Value             string    `json:"value"`
	Token             string    `json:"token,omitempty"`
	TokenSymbol       string    `json:"token_symbol,omitempty"`
	Type              string    `json:"type"`              // "transfer", "swap", "contract_call"
	Status            string    `json:"status"`            // "confirmed", "failed"
	GasUsed           string    `json:"gas_used,omitempty"`
	GasPrice          string    `json:"gas_price,omitempty"`
	TransactionFee    string    `json:"transaction_fee"`
	Timestamp         time.Time `json:"timestamp"`
	Confirmations     uint64    `json:"confirmations"`
	
	// Contract interaction details
	ContractAddress   string    `json:"contract_address,omitempty"`
	MethodName        string    `json:"method_name,omitempty"`
	InputData         string    `json:"input_data,omitempty"`
	
	// Token transfer details
	TokenTransfers    []TokenTransfer `json:"token_transfers,omitempty"`
}

// TokenTransfer represents an ERC-20 or SPL token transfer within a transaction
type TokenTransfer struct {
	From          string `json:"from"`
	To            string `json:"to"`
	Value         string `json:"value"`
	TokenAddress  string `json:"token_address"`
	TokenSymbol   string `json:"token_symbol"`
	TokenDecimals int    `json:"token_decimals"`
}

// TransactionHistoryFilter defines filtering options for transaction history queries
type TransactionHistoryFilter struct {
	Address    string
	FromBlock  *uint64
	ToBlock    *uint64
	Limit      int
	Offset     int
}