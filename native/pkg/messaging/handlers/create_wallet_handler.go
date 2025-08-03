// SPDX-License-Identifier: Apache-2.0
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/messaging"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"go.uber.org/zap"
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
func CreateCreateWalletHandler(walletManager wallet.IWalletManager, logger *zap.Logger) messaging.RpcHandler {
	return func(request messaging.RpcRequest) (messaging.RpcResponse, error) {
		logger.Info("CreateWallet RPC handler started", 
			zap.String("method", "create_wallet"),
			zap.String("request_id", request.ID))
		
		// Parse parameters
		var params CreateWalletParams
		if request.Params != nil {
			if err := json.Unmarshal(request.Params, &params); err != nil {
				logger.Error("Failed to parse create_wallet parameters", 
					zap.Error(err),
					zap.String("request_id", request.ID))
				return messaging.RpcResponse{
					Error: &messaging.ErrorInfo{
						Code:    -32602,
						Message: fmt.Sprintf("Invalid params: %s", err.Error()),
					},
				}, nil
			}
		}
		
		logger.Info("CreateWallet parameters parsed", 
			zap.String("chain", params.Chain),
			zap.Bool("has_password", params.Password != ""),
			zap.Int("password_length", len(params.Password)),
			zap.String("request_id", request.ID))

		// Validate required parameters
		if params.Chain == "" {
			logger.Error("CreateWallet validation failed: missing chain", 
				zap.String("request_id", request.ID))
			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    ErrInvalidChain,
					Message: "Chain is required",
				},
			}, nil
		}

		if params.Password == "" {
			logger.Error("CreateWallet validation failed: missing password", 
				zap.String("request_id", request.ID))
			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    ErrPasswordRequired,
					Message: "Password is required",
				},
			}, nil
		}

		logger.Info("CreateWallet validation passed, calling wallet manager", 
			zap.String("chain", params.Chain),
			zap.String("request_id", request.ID))

		// Create wallet using wallet manager
		address, publicKey, mnemonic, err := walletManager.CreateWallet(
			context.Background(),
			params.Chain,
			params.Password,
		)

		if err != nil {
			logger.Error("WalletManager.CreateWallet failed", 
				zap.Error(err),
				zap.String("chain", params.Chain),
				zap.String("request_id", request.ID))
		} else {
			logger.Info("WalletManager.CreateWallet succeeded", 
				zap.String("address", address),
				zap.String("public_key", publicKey[:20]+"..."), // Truncate for security
				zap.Int("mnemonic_word_count", len(strings.Fields(mnemonic))),
				zap.String("request_id", request.ID))
		}

		if err != nil {
			// Map error to specific error codes
			errorCode := -32000 // Default server error
			errorMessage := err.Error()

			switch {
			case strings.Contains(errorMessage, "invalid chain"):
				errorCode = ErrInvalidChain
			case strings.Contains(errorMessage, "wallet creation failed"):
				errorCode = ErrWalletCreationFailed
			}

			logger.Error("CreateWallet returning error response", 
				zap.Int("error_code", errorCode),
				zap.String("error_message", errorMessage),
				zap.String("request_id", request.ID))

			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    errorCode,
					Message: errorMessage,
				},
			}, nil
		}

		// Return the generated mnemonic along with wallet details
		result := CreateWalletResult{
			Address:   address,
			PublicKey: publicKey,
			Mnemonic:  mnemonic,
			CreatedAt: time.Now().Unix(),
		}

		logger.Info("CreateWallet preparing result", 
			zap.String("result_address", result.Address),
			zap.String("result_public_key", result.PublicKey[:20]+"..."),
			zap.Int("result_mnemonic_words", len(strings.Fields(result.Mnemonic))),
			zap.Int64("result_created_at", result.CreatedAt),
			zap.String("request_id", request.ID))

		resultJSON, err := json.Marshal(result)
		if err != nil {
			logger.Error("CreateWallet failed to marshal result", 
				zap.Error(err),
				zap.String("request_id", request.ID))
			return messaging.RpcResponse{
				Error: &messaging.ErrorInfo{
					Code:    -32000,
					Message: fmt.Sprintf("Failed to marshal result: %s", err.Error()),
				},
			}, nil
		}

		logger.Info("CreateWallet returning success response", 
			zap.Int("result_json_length", len(resultJSON)),
			zap.String("request_id", request.ID))

		return messaging.RpcResponse{
			Result: resultJSON,
		}, nil
	}
}