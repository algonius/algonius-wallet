package jito

import (
	"encoding/json"

	jitorpc "github.com/jito-labs/jito-go-rpc"
)

// IJitoAPI defines the interface for Jito API operations
type IJitoAPI interface {
	GetTipAccounts() (json.RawMessage, error)
	GetRandomTipAccount() (*jitorpc.TipAccount, error)

	GetBundleStatuses(bundleIds []string) (*jitorpc.BundleStatusResponse, error)
	SendBundle(params [][]string) (json.RawMessage, error) // Fixed signature to match jito-go-rpc
	GetInflightBundleStatuses(params interface{}) (json.RawMessage, error)

	SendTxn(params interface{}, bundleOnly bool) (json.RawMessage, error)
}