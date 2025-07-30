package okex

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

const (
	MaxSlippage        = 1.0
	MinSlippage        = 0.0
	MaxFeePercent      = 3.0
	MinFeePercent      = 0.0
	DefaultGasLevel    = "average"
	CallDataMemoLength = 128

	DefaultPriceImpactProtection = 0.9
	MinPriceImpactProtection     = 0.0
	MaxPriceImpactProtection     = 1.0

	TxStatusPending = "1"
	TxStatusSuccess = "2"
	TxStatusFailed  = "3"

	DefaultOrderLimit = "20"
	MaxOrderLimit     = "100"
)

// CommonResponse represents the common response structure
type CommonResponse[T any] struct {
	Code    string `json:"code"`
	Message string `json:"msg"`
	Data    T      `json:"data"`
}

// BroadcastTransactionParams represents parameters for broadcasting a transaction
type BroadcastTransactionParams struct {
	AccountID  string `json:"accountId"`
	ChainIndex string `json:"chainIndex"`
	Address    string `json:"address"`
	SignedTx   string `json:"signedTx"`
}

// Validate checks if the broadcast transaction parameters are valid
func (p BroadcastTransactionParams) Validate() error {
	if p.AccountID == "" && p.Address == "" {
		return fmt.Errorf("accountId or address is required")
	}

	if p.ChainIndex == "" {
		return fmt.Errorf("chainIndex is required")
	}

	if p.SignedTx == "" {
		return fmt.Errorf("signedTx is required")
	}

	if !strings.HasPrefix(p.SignedTx, "0x") {
		return fmt.Errorf("signedTx must start with 0x")
	}

	return nil
}

// BroadcastTransactionData represents the data field in broadcast transaction response
type BroadcastTransactionData struct {
	OrderId string `json:"orderId"`
}

// QueryOrdersParams represents parameters for querying broadcast transaction orders
type QueryOrdersParams struct {
	Address    string `json:"address,omitempty"`
	AccountID  string `json:"accountId,omitempty"`
	ChainIndex string `json:"chainIndex,omitempty"`
	TxStatus   string `json:"txStatus,omitempty"`
	OrderID    string `json:"orderId,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
	Limit      string `json:"limit,omitempty"`
}

// Validate checks if the query orders parameters are valid
func (p QueryOrdersParams) Validate() error {
	if p.Address == "" && p.AccountID == "" {
		return fmt.Errorf("either address or accountId is required")
	}

	if p.TxStatus != "" && p.TxStatus != TxStatusPending &&
		p.TxStatus != TxStatusSuccess && p.TxStatus != TxStatusFailed {
		return fmt.Errorf("invalid txStatus value")
	}

	if p.Limit != "" {
		limit, err := strconv.Atoi(p.Limit)
		if err != nil {
			return fmt.Errorf("invalid limit format")
		}
		if limit > 100 {
			return fmt.Errorf("limit cannot exceed 100")
		}
	}

	return nil
}

// OrderData represents a single order in the response
type OrderData struct {
	ChainIndex string `json:"chainIndex"`
	Address    string `json:"address"`
	AccountID  string `json:"accountId"`
	OrderID    string `json:"orderId"`
	TxStatus   string `json:"txStatus"`
	TxHash     string `json:"txHash"`
	Limit      string `json:"limit"`
}

// ChainData represents supported chain information
type ChainData struct {
	ChainID                int32  `json:"chainId"`
	ChainName              string `json:"chainName"`
	DexTokenApproveAddress string `json:"dexTokenApproveAddress"`
}

// ApproveTransactionParams represents parameters for approve transaction request
type ApproveTransactionParams struct {
	ChainID              string `json:"chainId"`
	TokenContractAddress string `json:"tokenContractAddress"`
	ApproveAmount        string `json:"approveAmount"`
}

// Validate checks if the approve transaction parameters are valid
func (p ApproveTransactionParams) Validate() error {
	if p.ChainID == "" {
		return fmt.Errorf("chainId is required")
	}
	if p.TokenContractAddress == "" {
		return fmt.Errorf("tokenContractAddress is required")
	}
	if p.ApproveAmount == "" {
		return fmt.Errorf("approveAmount is required")
	}
	return nil
}

// ApproveTransactionData represents the data field in approve transaction response
type ApproveTransactionData struct {
	Data               string `json:"data"`
	DexContractAddress string `json:"dexContractAddress"`
	GasLimit           string `json:"gasLimit"`
	GasPrice           string `json:"gasPrice"`
}

// Response type definitions
type BroadcastTransactionResponse = CommonResponse[[]BroadcastTransactionData]
type QueryOrdersResponse = CommonResponse[[]OrderData]
type SupportedChainResponse = CommonResponse[[]ChainData]
type ApproveTransactionResponse = CommonResponse[ApproveTransactionData]

// IOKEXClient defines the interface for OKEX API operations (wallet-focused methods)
type IOKEXClient interface {
	BroadcastTransaction(ctx context.Context, params BroadcastTransactionParams) (*BroadcastTransactionResponse, error)
	GetOrders(ctx context.Context, params QueryOrdersParams) (*QueryOrdersResponse, error)
	GetSupportedChains(ctx context.Context, chainID string) (*SupportedChainResponse, error)
	GetApproveTransaction(ctx context.Context, params ApproveTransactionParams) (*ApproveTransactionResponse, error)
}