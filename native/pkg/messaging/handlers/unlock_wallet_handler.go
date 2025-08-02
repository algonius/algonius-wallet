// SPDX-License-Identifier: Apache-2.0
package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/algonius/algonius-wallet/native/pkg/messaging"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"go.uber.org/zap"
)

// UnlockWalletParams represents the parameters for unlock_wallet RPC method
type UnlockWalletParams struct {
	Password string `json:"password"`
}

// UnlockWalletResult represents the result of unlock_wallet RPC method
type UnlockWalletResult struct {
	Address   string          `json:"address"`
	PublicKey string          `json:"public_key"`
	Chains    map[string]bool `json:"chains"`
	UnlockedAt int64          `json:"unlocked_at"`
}

// WalletStatusResult represents the result of wallet status check
type WalletStatusResult struct {
	HasWallet  bool   `json:"hasWallet"`
	IsUnlocked bool   `json:"isUnlocked"`
	Address    string `json:"address,omitempty"`
}

// CreateUnlockWalletHandler creates an RPC handler for unlock_wallet method
func CreateUnlockWalletHandler(walletManager wallet.IWalletManager) messaging.RpcHandler {
	return func(request messaging.RpcRequest) (messaging.RpcResponse, error) {
		
		// Parse parameters
		var params UnlockWalletParams
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
		if params.Password == "" {
			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    -32602,
					Message: "Password is required",
				},
			}, nil
		}

		// Check if wallet exists
		if !walletManager.HasWallet() {
			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    -32004,
					Message: "No wallet found. Please create or import a wallet first.",
				},
			}, nil
		}

		// Unlock wallet
		err := walletManager.UnlockWallet(params.Password)
		if err != nil {
			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    -32001,
					Message: fmt.Sprintf("Failed to unlock wallet: %s", err.Error()),
				},
			}, nil
		}

		// Get current wallet status
		walletStatus := walletManager.GetCurrentWallet()
		if walletStatus == nil {
			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    -32000,
					Message: "Failed to get wallet status after unlock",
				},
			}, nil
		}

		// Prepare result
		result := UnlockWalletResult{
			Address:    walletStatus.Address,
			PublicKey:  walletStatus.PublicKey,
			Chains:     walletStatus.Chains,
			UnlockedAt: walletStatus.LastUsed,
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

// CreateLockWalletHandler creates an RPC handler for lock_wallet method
func CreateLockWalletHandler(walletManager wallet.IWalletManager) messaging.RpcHandler {
	return func(request messaging.RpcRequest) (messaging.RpcResponse, error) {
		// Lock the wallet
		walletManager.LockWallet()

		// Return success
		result := map[string]interface{}{
			"locked": true,
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

// CreateWalletStatusHandler creates an RPC handler for wallet_status method
func CreateWalletStatusHandler(walletManager wallet.IWalletManager, logger *zap.Logger) messaging.RpcHandler {
	return func(request messaging.RpcRequest) (messaging.RpcResponse, error) {
		// Check wallet status
		hasWallet := walletManager.HasWallet()
		isUnlocked := walletManager.IsUnlocked()
		
		result := WalletStatusResult{
			HasWallet:  hasWallet,
			IsUnlocked: isUnlocked,
		}
		
		// Add address if wallet is unlocked
		if result.IsUnlocked {
			walletStatus := walletManager.GetCurrentWallet()
			if walletStatus != nil {
				result.Address = walletStatus.Address
			}
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