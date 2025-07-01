// SPDX-License-Identifier: Apache-2.0
package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/algonius/algonius-wallet/native/pkg/messaging"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
)

// ImportWalletParams represents the parameters for import_wallet RPC method
type ImportWalletParams struct {
	Mnemonic       string `json:"mnemonic"`
	Password       string `json:"password"`
	Chain          string `json:"chain"`
	DerivationPath string `json:"derivation_path,omitempty"`
}

// ImportWalletResult represents the result of import_wallet RPC method
type ImportWalletResult struct {
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	ImportedAt int64  `json:"imported_at"`
}

// ImportWalletError codes as specified in the requirements
const (
	ErrInvalidMnemonic      = -32001
	ErrWeakPassword         = -32002
	ErrUnsupportedChain     = -32003
	ErrWalletAlreadyExists  = -32004
	ErrStorageEncryptionFailed = -32005
)

// CreateImportWalletHandler creates an RPC handler for import_wallet method
func CreateImportWalletHandler(walletManager wallet.IWalletManager) messaging.RpcHandler {
	return func(request messaging.RpcRequest) (messaging.RpcResponse, error) {
		// Parse parameters
		var params ImportWalletParams
		if request.Params != nil {
			if err := json.Unmarshal(request.Params, &params); err != nil {
				return messaging.RpcResponse{
					Error: &messaging.ErrorInfo{
						Code:    -32602,
						Message: fmt.Sprintf("Invalid params: %s", err.Error()),
					},
				}, nil
			}
		}

		// Validate required parameters
		if params.Mnemonic == "" {
			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    ErrInvalidMnemonic,
					Message: "Mnemonic phrase is required",
				},
			}, nil
		}

		if params.Password == "" {
			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    ErrWeakPassword,
					Message: "Password is required",
				},
			}, nil
		}

		if params.Chain == "" {
			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    ErrUnsupportedChain,
					Message: "Chain is required",
				},
			}, nil
		}

		// Import wallet using wallet manager
		address, publicKey, importedAt, err := walletManager.ImportWallet(
			context.Background(),
			params.Mnemonic,
			params.Password,
			params.Chain,
			params.DerivationPath,
		)

		if err != nil {
			// Map error to specific error codes
			errorCode := -32000 // Default server error
			errorMessage := err.Error()

			switch {
			case contains(errorMessage, "invalid mnemonic"):
				errorCode = ErrInvalidMnemonic
			case contains(errorMessage, "weak password"):
				errorCode = ErrWeakPassword
			case contains(errorMessage, "unsupported chain"):
				errorCode = ErrUnsupportedChain
			case contains(errorMessage, "wallet already exists"):
				errorCode = ErrWalletAlreadyExists
			case contains(errorMessage, "storage encryption failed"):
				errorCode = ErrStorageEncryptionFailed
			}

			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    errorCode,
					Message: errorMessage,
				},
			}, nil
		}

		// Prepare result
		result := ImportWalletResult{
			Address:    address,
			PublicKey:  publicKey,
			ImportedAt: importedAt,
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    -32000,
					Message: fmt.Sprintf("Failed to marshal result: %s", err.Error()),
				},
			}, nil
		}

		return messaging.RpcResponse{
			Result: resultJSON,
		}, nil
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && containsAt(s, substr, 0)))
}

// containsAt checks if substr is found in s starting at any position
func containsAt(s, substr string, start int) bool {
	if start > len(s)-len(substr) {
		return false
	}
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}