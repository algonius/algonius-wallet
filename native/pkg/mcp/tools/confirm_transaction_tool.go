// Package tools provides MCP tool implementations for the Algonius Native Host.
package tools

import (
	"context"
	"fmt"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/algonius/algonius-wallet/native/pkg/wallet/chain"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ConfirmTransactionTool implements the MCP "confirm_transaction" tool for checking transaction status.
type ConfirmTransactionTool struct {
	manager wallet.IWalletManager
	factory *chain.ChainFactory
}

// NewConfirmTransactionTool constructs a ConfirmTransactionTool with the given wallet manager.
func NewConfirmTransactionTool(manager wallet.IWalletManager) *ConfirmTransactionTool {
	return &ConfirmTransactionTool{
		manager: manager,
		factory: chain.NewChainFactory(),
	}
}

// GetMeta returns the MCP tool definition for "confirm_transaction" as per the documented API schema.
func (t *ConfirmTransactionTool) GetMeta() mcp.Tool {
	return mcp.NewTool("confirm_transaction",
		mcp.WithDescription("Check transaction confirmation status"),
		mcp.WithString("chain",
			mcp.Required(),
			mcp.Description("Chain identifier (ethereum, bsc)"),
		),
		mcp.WithString("tx_hash",
			mcp.Required(),
			mcp.Description("Transaction hash"),
		),
		mcp.WithNumber("required_confirmations",
			mcp.Description("Required confirmations for finality (default: 6 for Ethereum, 3 for BSC)"),
		),
	)
}

// GetHandler returns the handler function for the "confirm_transaction" tool.
// The handler checks the confirmation status of the specified transaction.
func (t *ConfirmTransactionTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract and validate chain parameter
		chain, err := req.RequireString("chain")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid 'chain' parameter"), nil
		}

		// Normalize chain name
		normalizedChain, err := t.normalizeChainName(chain)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("unsupported chain: %s", chain)), nil
		}

		// Extract and validate tx_hash parameter
		txHash, err := req.RequireString("tx_hash")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid 'tx_hash' parameter"), nil
		}

		// Extract optional required_confirmations parameter
		var requiredConfirmations uint64
		requiredConfirmations = uint64(req.GetFloat("required_confirmations", 0))

		// Check transaction confirmation status
		confirmation, err := wallet.ConfirmTransaction(ctx, normalizedChain, txHash, requiredConfirmations, t.factory)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to check transaction confirmation: %s", err.Error())), nil
		}

		// Prepare response as markdown text
		markdown := fmt.Sprintf("### Transaction Confirmation Status\n\n"+
			"- **Transaction Hash**: `%s`\n"+
			"- **Chain**: `%s`\n"+
			"- **Status**: `%s`\n"+
			"- **Confirmations**: `%d`\n"+
			"- **Required Confirmations**: `%d`\n"+
			"- **Block Number**: `%d`\n"+
			"- **Gas Used**: `%s`\n"+
			"- **Transaction Fee**: `%s`\n"+
			"- **Timestamp**: `%s`\n",
			confirmation.TxHash,
			normalizedChain,
			confirmation.Status,
			confirmation.Confirmations,
			confirmation.RequiredConfirmations,
			confirmation.BlockNumber,
			confirmation.GasUsed,
			confirmation.TransactionFee,
			confirmation.Timestamp.Format("2006-01-02T15:04:05Z"),
		)

		return mcp.NewToolResultText(markdown), nil
	}
}

// normalizeChainName converts various chain name formats to standard form
func (t *ConfirmTransactionTool) normalizeChainName(chain string) (string, error) {
	switch chain {
	case "ethereum", "eth", "ETH", "Ethereum":
		return "ETH", nil
	case "bsc", "BSC", "binance", "Binance", "binance-smart-chain":
		return "BSC", nil
	default:
		return "", fmt.Errorf("unsupported chain: %s", chain)
	}
}
