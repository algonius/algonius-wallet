// SPDX-License-Identifier: Apache-2.0
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/messaging"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
)

// CreateWalletParams represents the parameters for create_wallet RPC method
type CreateWalletParams struct {
	Chain    string `json:"chain"`
	Password string `json:"password"`
}

// CreateWalletResult represents the result of create_wallet RPC method
type CreateWalletResult struct {
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
	Mnemonic  string `json:"mnemonic"`
	CreatedAt int64  `json:"created_at"`
}

// CreateWalletError codes
const (
	ErrInvalidChain         = -33001
	ErrWalletCreationFailed = -33002
	ErrPasswordRequired     = -33003
)

// CreateCreateWalletHandler creates an RPC handler for create_wallet method
func CreateCreateWalletHandler(walletManager wallet.IWalletManager) messaging.RpcHandler {
	return func(request messaging.RpcRequest) (messaging.RpcResponse, error) {
		// Parse parameters
		var params CreateWalletParams
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
		if params.Chain == "" {
			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    ErrInvalidChain,
					Message: "Chain is required",
				},
			}, nil
		}

		if params.Password == "" {
			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    ErrPasswordRequired,
					Message: "Password is required",
				},
			}, nil
		}

		// Create wallet using wallet manager
		address, publicKey, err := walletManager.CreateWallet(
			context.Background(),
			params.Chain,
		)

		if err != nil {
			// Map error to specific error codes
			errorCode := -32000 // Default server error
			errorMessage := err.Error()

			switch {
			case contains(errorMessage, "invalid chain"):
				errorCode = ErrInvalidChain
			case contains(errorMessage, "wallet creation failed"):
				errorCode = ErrWalletCreationFailed
			}

			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    errorCode,
					Message: errorMessage,
				},
			}, nil
		}

		// For now, we'll return an empty mnemonic since the current CreateWallet method
		// doesn't generate or return a mnemonic. This should be updated when the
		// wallet manager is enhanced to support mnemonic generation.
		result := CreateWalletResult{
			Address:   address,
			PublicKey: publicKey,
			Mnemonic:  "", // TODO: Add mnemonic generation to wallet manager
			CreatedAt: time.Now().Unix(),
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