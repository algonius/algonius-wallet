// Package tools provides MCP tool implementations for the Algonius Native Host.
package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/algonius/algonius-wallet/native/pkg/errors"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/toolutils"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/algonius/algonius-wallet/native/pkg/wallet/chain"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// GetTransactionStatusTool implements the MCP "get_transaction_status" tool for checking blockchain transaction status.
type GetTransactionStatusTool struct {
	manager wallet.IWalletManager
	logger  *zap.Logger
}

// NewGetTransactionStatusTool constructs a GetTransactionStatusTool with the given wallet manager.
func NewGetTransactionStatusTool(manager wallet.IWalletManager, logger *zap.Logger) *GetTransactionStatusTool {
	if logger == nil {
		logger = zap.NewNop() // Use no-op logger if none provided
	}
	return &GetTransactionStatusTool{
		manager: manager,
		logger:  logger,
	}
}

// GetMeta returns the MCP tool definition for "get_transaction_status" as per the documented API schema.
func (t *GetTransactionStatusTool) GetMeta() mcp.Tool {
	return mcp.NewTool("get_transaction_status",
		mcp.WithDescription("Get the current status of a blockchain transaction by its hash"),
		mcp.WithString("transaction_hash",
			mcp.Required(),
			mcp.Description("The hash of the transaction to check"),
		),
		mcp.WithString("chain",
			mcp.Description("The blockchain network (optional, will try to detect if not provided)"),
		),
	)
}

// GetHandler returns the handler function for the "get_transaction_status" tool.
// The handler checks the status of a transaction on the blockchain.
func (t *GetTransactionStatusTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract and validate transaction_hash parameter
		txHash, err := req.RequireString("transaction_hash")
		if err != nil {
			toolErr := errors.MissingRequiredFieldError("transaction_hash")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Validate transaction hash format
		if txHash == "" {
			toolErr := errors.ValidationError("transaction_hash", "transaction hash cannot be empty")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Extract optional chain parameter
		chainName := req.GetString("chain", "")

		t.logger.Debug("Checking transaction status",
			zap.String("transaction_hash", txHash),
			zap.String("chain", chainName))

		// If chain not provided, try to detect it
		if chainName == "" {
			chainName = t.detectChainFromHash(txHash)
			if chainName == "" {
				toolErr := errors.ValidationError("chain", "unable to detect chain from transaction hash, please specify chain parameter")
				return toolutils.FormatErrorResult(toolErr), nil
			}
			t.logger.Debug("Detected chain from transaction hash",
				zap.String("transaction_hash", txHash),
				zap.String("detected_chain", chainName))
		}

		// Get chain interface
		chainInterface, err := t.getChainInterface(chainName)
		if err != nil {
			toolErr := errors.ValidationError("chain", fmt.Sprintf("unsupported chain: %s", chainName))
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Check transaction status on blockchain
		confirmation, err := chainInterface.ConfirmTransaction(ctx, txHash, 1) // Require at least 1 confirmation
		if err != nil {
			t.logger.Error("Failed to check transaction status",
				zap.String("transaction_hash", txHash),
				zap.String("chain", chainName),
				zap.Error(err))
			
			// Try to determine if it's a "not found" error
			if strings.Contains(strings.ToLower(err.Error()), "not found") || 
			   strings.Contains(strings.ToLower(err.Error()), "does not exist") {
				markdown := fmt.Sprintf("### Transaction Status: Not Found ❓\n\n"+
					"- **Transaction Hash**: `%s`\n"+
					"- **Chain**: `%s`\n"+
					"- **Status**: `not_found`\n"+
					"- **Details**: Transaction hash not found on the blockchain\n",
					txHash, chainName)
				return mcp.NewToolResultText(markdown), nil
			}
			
			toolErr := errors.InternalError("check transaction status", err)
			return toolutils.FormatErrorResult(toolErr), nil
		}

		var markdown string
		if confirmation.Status == "confirmed" {
			markdown = fmt.Sprintf("### Transaction Status: Confirmed ✅\n\n"+
				"- **Transaction Hash**: `%s`\n"+
				"- **Chain**: `%s`\n"+
				"- **Status**: `confirmed`\n"+
				"- **Confirmations**: `%d`\n"+
				"- **Block Number**: `%d`\n"+
				"- **Gas Used**: `%d`\n"+
				"- **Transaction Fee**: `%s`\n"+
				"- **Timestamp**: `%s`\n",
				txHash, chainName, confirmation.Confirmations, confirmation.BlockNumber,
				confirmation.GasUsed, confirmation.TransactionFee, confirmation.Timestamp.Format("2006-01-02 15:04:05 UTC"))
		} else if confirmation.Status == "pending" {
			markdown = fmt.Sprintf("### Transaction Status: Pending ⏳\n\n"+
				"- **Transaction Hash**: `%s`\n"+
				"- **Chain**: `%s`\n"+
				"- **Status**: `pending`\n"+
				"- **Details**: Transaction is waiting to be included in a block\n",
				txHash, chainName)
		} else if confirmation.Status == "failed" {
			markdown = fmt.Sprintf("### Transaction Status: Failed ❌\n\n"+
				"- **Transaction Hash**: `%s`\n"+
				"- **Chain**: `%s`\n"+
				"- **Status**: `failed`\n"+
				"- **Error**: `%s`\n"+
				"- **Timestamp**: `%s`\n",
				txHash, chainName, confirmation.ErrorMessage, confirmation.Timestamp.Format("2006-01-02 15:04:05 UTC"))
		} else {
			markdown = fmt.Sprintf("### Transaction Status: %s\n\n"+
				"- **Transaction Hash**: `%s`\n"+
				"- **Chain**: `%s`\n"+
				"- **Status**: `%s`\n"+
				"- **Details**: Transaction status is unknown or in an unexpected state\n",
				strings.Title(confirmation.Status), txHash, chainName, confirmation.Status)
		}

		return mcp.NewToolResultText(markdown), nil
	}
}

// detectChainFromHash attempts to determine the chain based on the transaction hash format
func (t *GetTransactionStatusTool) detectChainFromHash(txHash string) string {
	// Ethereum-style hashes start with 0x and are 66 characters long (0x + 64 hex chars)
	if len(txHash) == 66 && strings.HasPrefix(txHash, "0x") {
		return "ethereum"
	}
	
	// Solana-style hashes are base58 encoded and typically 88 characters long
	// This is a simple heuristic - a more robust implementation would validate base58
	if len(txHash) >= 80 && len(txHash) <= 90 && !strings.Contains(txHash, "0x") {
		return "solana"
	}
	
	// BSC also uses Ethereum-style hashes
	// We'll default to ethereum for 0x prefixed hashes
	// A more sophisticated implementation might need additional context
	
	return "" // Unable to detect
}

// getChainInterface gets the appropriate chain interface for the given chain name
func (t *GetTransactionStatusTool) getChainInterface(chainName string) (chain.Chain, error) {
	switch strings.ToLower(chainName) {
	case "solana", "sol":
		return chain.NewSolanaChainLegacy(), nil
	case "ethereum", "eth":
		return chain.NewETHChain(nil, t.logger), nil
	case "bsc", "binance smart chain":
		return chain.NewBSCChain(nil, t.logger), nil
	default:
		return nil, fmt.Errorf("unsupported chain: %s", chainName)
	}
}