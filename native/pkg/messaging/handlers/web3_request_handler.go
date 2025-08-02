// Package handlers provides Native Messaging handlers for the Algonius Native Host.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/event"
	"github.com/algonius/algonius-wallet/native/pkg/messaging"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
)

// Web3RequestParams represents the parameters for a web3 request from a web page
type Web3RequestParams struct {
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
	Origin string      `json:"origin,omitempty"`
}

// TransactionParams represents the parameters for an eth_sendTransaction request
type TransactionParams struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Value    string `json:"value,omitempty"`
	Gas      string `json:"gas,omitempty"`
	GasPrice string `json:"gasPrice,omitempty"`
	Data     string `json:"data,omitempty"`
	Nonce    string `json:"nonce,omitempty"`
}

// CreateWeb3RequestHandler creates a handler for web3 requests from web pages
func CreateWeb3RequestHandler(manager wallet.IWalletManager, broadcaster *event.EventBroadcaster) messaging.RpcHandler {
	return func(req messaging.RpcRequest) (messaging.RpcResponse, error) {
		var params Web3RequestParams
		if req.Params != nil {
			if err := json.Unmarshal(req.Params, &params); err != nil {
				return messaging.RpcResponse{
					ID: req.ID,
					Error: &messaging.ErrorInfo{
						Code:    -32602,
						Message: "Invalid web3 request params: " + err.Error(),
					},
				}, nil
			}
		}

		// Handle different Web3 methods
		switch params.Method {
		case "eth_requestAccounts":
			return handleRequestAccounts(req.ID, manager)
		
		case "eth_accounts":
			return handleGetAccounts(req.ID, manager)
		
		case "eth_chainId":
			return handleGetChainId(req.ID)
		
		case "eth_sendTransaction":
			return handleSendTransaction(req.ID, params, manager, broadcaster)
		
		case "personal_sign":
			return handlePersonalSign(req.ID, params, manager, broadcaster)
		
		case "signMessage":
			return handleSolanaSignMessage(req.ID, params, manager, broadcaster)
		
		default:
			return messaging.RpcResponse{
				ID: req.ID,
				Error: &messaging.ErrorInfo{
					Code:    -32601,
					Message: fmt.Sprintf("Method %s not supported", params.Method),
				},
			}, nil
		}
	}
}

// handleRequestAccounts handles eth_requestAccounts requests
func handleRequestAccounts(id string, manager wallet.IWalletManager) (messaging.RpcResponse, error) {
	// Get available accounts from wallet manager
	ctx := context.Background()
	accounts, err := manager.GetAccounts(ctx)
	if err != nil {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32000,
				Message: "Failed to get accounts: " + err.Error(),
			},
		}, nil
	}

	result, _ := json.Marshal(accounts)
	return messaging.RpcResponse{
		ID:     id,
		Result: result,
	}, nil
}

// handleGetAccounts handles eth_accounts requests
func handleGetAccounts(id string, manager wallet.IWalletManager) (messaging.RpcResponse, error) {
	return handleRequestAccounts(id, manager)
}

// handleGetChainId handles eth_chainId requests
func handleGetChainId(id string) (messaging.RpcResponse, error) {
	// Return Ethereum mainnet by default (0x1)
	// TODO: Make this configurable based on user preferences
	result, _ := json.Marshal("0x1")
	return messaging.RpcResponse{
		ID:     id,
		Result: result,
	}, nil
}

// handleSendTransaction handles eth_sendTransaction requests from web pages
func handleSendTransaction(id string, params Web3RequestParams, manager wallet.IWalletManager, broadcaster *event.EventBroadcaster) (messaging.RpcResponse, error) {
	// Parse transaction parameters
	var txParams []TransactionParams
	paramsBytes, err := json.Marshal(params.Params)
	if err != nil {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32602,
				Message: "Invalid transaction params: " + err.Error(),
			},
		}, nil
	}
	
	if err := json.Unmarshal(paramsBytes, &txParams); err != nil {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32602,
				Message: "Invalid transaction params format: " + err.Error(),
			},
		}, nil
	}
	
	if len(txParams) == 0 {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32602,
				Message: "Missing transaction parameters",
			},
		}, nil
	}
	
	txParam := txParams[0]

	// Create pending transaction
	ctx := context.Background()
	pendingTx := &wallet.PendingTransaction{
		Hash:                      generateTransactionHash(), // Generate temporary hash
		Chain:                     "ethereum", // Default to Ethereum for eth_sendTransaction
		From:                      txParam.From,
		To:                        txParam.To,
		Amount:                    txParam.Value,
		Token:                     "ETH",
		Type:                      "transfer",
		Status:                    "pending",
		Confirmations:             0,
		RequiredConfirmations:     6,
		GasFee:                    txParam.Gas,
		Priority:                  "medium",
		EstimatedConfirmationTime: "2-5 minutes",
		SubmittedAt:               time.Now(),
		LastChecked:               time.Now(),
	}

	// Add transaction to pending queue
	if err := manager.AddPendingTransaction(ctx, pendingTx); err != nil {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32000,
				Message: "Failed to add pending transaction: " + err.Error(),
			},
		}, nil
	}

	// Broadcast transaction confirmation needed event to AI Agent
	if broadcaster != nil {
		event := &event.Event{
			Type: "transaction_confirmation_needed",
			Data: map[string]interface{}{
				"transaction_hash": pendingTx.Hash,
				"chain":           pendingTx.Chain,
				"from":            pendingTx.From,
				"to":              pendingTx.To,
				"amount":          pendingTx.Amount,
				"token":           pendingTx.Token,
				"origin":          params.Origin,
				"gas_fee":         pendingTx.GasFee,
				"submitted_at":    pendingTx.SubmittedAt.Format(time.RFC3339),
			},
		}
		broadcaster.Broadcast(event)
	}

	// Return the pending transaction hash
	result, _ := json.Marshal(pendingTx.Hash)
	return messaging.RpcResponse{
		ID:     id,
		Result: result,
	}, nil
}

// handlePersonalSign handles personal_sign requests from web pages
func handlePersonalSign(id string, params Web3RequestParams, manager wallet.IWalletManager, broadcaster *event.EventBroadcaster) (messaging.RpcResponse, error) {
	// Parse signing parameters
	var signParams []interface{}
	paramsBytes, err := json.Marshal(params.Params)
	if err != nil {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32602,
				Message: "Invalid signing params: " + err.Error(),
			},
		}, nil
	}
	
	if err := json.Unmarshal(paramsBytes, &signParams); err != nil {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32602,
				Message: "Invalid signing params format: " + err.Error(),
			},
		}, nil
	}
	
	if len(signParams) < 2 {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32602,
				Message: "Missing signing parameters",
			},
		}, nil
	}
	
	// Extract message and address
	var message string
	var address string
	
	// Handle different parameter types for the message
	switch msg := signParams[0].(type) {
	case string:
		message = msg
	case []byte:
		message = string(msg)
	case []interface{}:
		// Convert byte array to string
		// Check if this is a map with numeric keys (which is how JS objects are serialized)
		if len(msg) > 0 {
			// Regular array of numbers
			byteArray := make([]byte, len(msg))
			for i, v := range msg {
				if val, ok := v.(float64); ok {
					byteArray[i] = byte(val)
				}
			}
			message = string(byteArray)
		}
	case map[string]interface{}:
		// Handle map representation of byte array
		// Find the length of the map to determine array size
		maxIndex := -1
		for key := range msg {
			if idx, err := strconv.Atoi(key); err == nil && idx > maxIndex {
				maxIndex = idx
			}
		}
		
		if maxIndex >= 0 {
			byteArray := make([]byte, maxIndex+1)
			for key, value := range msg {
				if idx, err := strconv.Atoi(key); err == nil {
					if num, ok := value.(float64); ok {
						byteArray[idx] = byte(num)
					}
				}
			}
			message = string(byteArray)
		}
	default:
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32602,
				Message: fmt.Sprintf("Invalid message format: %T", msg),
			},
		}, nil
	}
	
	// Handle address parameter
	if addr, ok := signParams[1].(string); ok {
		address = addr
	} else {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32602,
				Message: "Invalid address format",
			},
		}, nil
	}

	// Sign the message using the wallet manager
	ctx := context.Background()
	signature, err := manager.SignMessage(ctx, address, message)
	if err != nil {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32000,
				Message: "Failed to sign message: " + err.Error(),
			},
		}, nil
	}

	// Return the signature
	result, _ := json.Marshal(signature)
	return messaging.RpcResponse{
		ID:     id,
		Result: result,
	}, nil
}

// handleSolanaSignMessage handles signMessage requests from Solana web pages
func handleSolanaSignMessage(id string, params Web3RequestParams, manager wallet.IWalletManager, broadcaster *event.EventBroadcaster) (messaging.RpcResponse, error) {
	// Parse signing parameters
	var signParams []interface{}
	paramsBytes, err := json.Marshal(params.Params)
	if err != nil {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32602,
				Message: "Invalid signing params: " + err.Error(),
			},
		}, nil
	}
	
	if err := json.Unmarshal(paramsBytes, &signParams); err != nil {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32602,
				Message: "Invalid signing params format: " + err.Error(),
			},
		}, nil
	}
	
	if len(signParams) < 1 {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32602,
				Message: "Missing signing parameters",
			},
		}, nil
	}
	
	// Extract message
	var messageBytes []byte
	
	// Handle different parameter types for the message
	switch msg := signParams[0].(type) {
	case string:
		// If it's a string, convert to bytes
		messageBytes = []byte(msg)
	case []byte:
		messageBytes = msg
	case []interface{}:
		// Convert array of numbers to byte array
		messageBytes = make([]byte, len(msg))
		for i, v := range msg {
			if val, ok := v.(float64); ok {
				messageBytes[i] = byte(val)
			}
		}
	case map[string]interface{}:
		// Handle map representation of byte array
		// Find the length of the map to determine array size
		maxIndex := -1
		for key := range msg {
			if idx, err := strconv.Atoi(key); err == nil && idx > maxIndex {
				maxIndex = idx
			}
		}
		
		if maxIndex >= 0 {
			messageBytes = make([]byte, maxIndex+1)
			for key, value := range msg {
				if idx, err := strconv.Atoi(key); err == nil {
					if num, ok := value.(float64); ok {
						messageBytes[idx] = byte(num)
					}
				}
			}
		}
	default:
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32602,
				Message: fmt.Sprintf("Invalid message format: %T", msg),
			},
		}, nil
	}

	// Get current accounts to find the signing address
	ctx := context.Background()
	accounts, err := manager.GetAccounts(ctx)
	if err != nil || len(accounts) == 0 {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32000,
				Message: "Failed to get accounts: " + err.Error(),
			},
		}, nil
	}

	// Use the first account as the signing address
	address := accounts[0]

	// For Solana, we need to convert the message bytes to a string representation
	// that our manager can work with
	message := string(messageBytes)

	// Sign the message using the wallet manager
	signature, err := manager.SignMessage(ctx, address, message)
	if err != nil {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32000,
				Message: "Failed to sign message: " + err.Error(),
			},
		}, nil
	}

	// For Solana, we need to return the signature in the correct format
	// The signature should be a base58 encoded string for Solana
	result, _ := json.Marshal(map[string]interface{}{
		"signature": signature,
		"publicKey": address,
	})
	
	return messaging.RpcResponse{
		ID:     id,
		Result: result,
	}, nil
}

// generateTransactionHash generates a temporary transaction hash for pending transactions
func generateTransactionHash() string {
	return fmt.Sprintf("0x%x", time.Now().UnixNano())
}