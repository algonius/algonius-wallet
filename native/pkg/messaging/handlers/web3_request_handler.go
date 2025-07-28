// Package handlers provides Native Messaging handlers for the Algonius Native Host.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/errors"
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
	var signParams []string
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
	
	message := signParams[0]
	address := signParams[1]

	// Create pending signature request
	ctx := context.Background()
	pendingSig := &wallet.PendingTransaction{
		Hash:                      generateTransactionHash(), // Generate temporary hash for signature request
		Chain:                     "ethereum",
		From:                      address,
		To:                        address, // For signing, from and to are the same
		Amount:                    "0",
		Token:                     "ETH",
		Type:                      "sign_message",
		Status:                    "pending",
		Confirmations:             0,
		RequiredConfirmations:     1,
		GasFee:                    "0",
		Priority:                  "high",
		EstimatedConfirmationTime: "immediate",
		SubmittedAt:               time.Now(),
		LastChecked:               time.Now(),
	}

	// Store the message to be signed in the details
	pendingSig.RejectionDetails = message // Reuse this field for message content

	// Add signature request to pending queue
	if err := manager.AddPendingTransaction(ctx, pendingSig); err != nil {
		return messaging.RpcResponse{
			ID: id,
			Error: &messaging.ErrorInfo{
				Code:    -32000,
				Message: "Failed to add pending signature request: " + err.Error(),
			},
		}, nil
	}

	// Broadcast signature confirmation needed event to AI Agent
	if broadcaster != nil {
		event := &event.Event{
			Type: "signature_confirmation_needed",
			Data: map[string]interface{}{
				"request_hash": pendingSig.Hash,
				"address":      address,
				"message":      message,
				"origin":       params.Origin,
				"submitted_at": pendingSig.SubmittedAt.Format(time.RFC3339),
			},
		}
		broadcaster.Broadcast(event)
	}

	// Return the pending signature request hash
	result, _ := json.Marshal(pendingSig.Hash)
	return messaging.RpcResponse{
		ID:     id,
		Result: result,
	}, nil
}

// generateTransactionHash generates a temporary transaction hash for pending transactions
func generateTransactionHash() string {
	return fmt.Sprintf("0x%x", time.Now().UnixNano())
}