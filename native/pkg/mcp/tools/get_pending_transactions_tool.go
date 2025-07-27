// Package tools provides MCP tool implementations for the Algonius Native Host.
package tools

import (
	"context"
	
	"fmt"

	"github.com/algonius/algonius-wallet/native/pkg/errors"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/toolutils"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GetPendingTransactionsTool implements the MCP "get_pending_transactions" tool for querying pending transactions.
type GetPendingTransactionsTool struct {
	manager wallet.IWalletManager
}

// NewGetPendingTransactionsTool constructs a GetPendingTransactionsTool with the given wallet manager.
func NewGetPendingTransactionsTool(manager wallet.IWalletManager) *GetPendingTransactionsTool {
	return &GetPendingTransactionsTool{manager: manager}
}

// GetMeta returns the MCP tool definition for "get_pending_transactions" as per the documented API schema.
func (t *GetPendingTransactionsTool) GetMeta() mcp.Tool {
	return mcp.NewTool("get_pending_transactions",
		mcp.WithDescription("Query pending transactions that require confirmation"),
		mcp.WithString("chain",
			mcp.Description("Blockchain chain name (e.g., 'ethereum', 'bsc', 'solana'). Leave empty for all chains"),
		),
		mcp.WithString("address",
			mcp.Description("Filter by specific wallet address. Leave empty for all controlled addresses"),
		),
		mcp.WithString("type",
			mcp.Description("Filter by transaction type (e.g., 'transfer', 'swap', 'contract'). Leave empty for all types"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of transactions to return (default: 10, max: 100)"),
		),
		mcp.WithNumber("offset",
			mcp.Description("Number of transactions to skip for pagination (default: 0)"),
		),
	)
}

// GetHandler returns the handler function for the "get_pending_transactions" tool.
// The handler queries pending transactions with optional filtering and pagination.
func (t *GetPendingTransactionsTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse optional parameters
		chain := req.GetString("chain", "")
		address := req.GetString("address", "")
		transactionType := req.GetString("type", "")
		
		// Parse pagination parameters
		limit := int(req.GetFloat("limit", 10))
		if limit <= 0 {
			limit = 10 // Default limit
		}
		if limit > 100 {
			limit = 100 // Maximum limit
		}
		
		offset := int(req.GetFloat("offset", 0))
		if offset < 0 {
			offset = 0
		}

		// Get pending transactions from wallet manager
		pendingTxs, err := t.manager.GetPendingTransactions(ctx, chain, address, transactionType, limit, offset)
		if err != nil {
			toolErr := errors.InternalError("get pending transactions", err)
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Format response as markdown
		markdown := "### Pending Transactions\n\n"
		
		if len(pendingTxs) == 0 {
			markdown += "No pending transactions found.\n"
		} else {
			markdown += fmt.Sprintf("Found %d pending transactions:\n\n", len(pendingTxs))
			
			for i, tx := range pendingTxs {
				markdown += fmt.Sprintf("#### Transaction %d\n", i+1)
				markdown += fmt.Sprintf("- **Hash**: `%s`\n", tx.Hash)
				markdown += fmt.Sprintf("- **Chain**: `%s`\n", tx.Chain)
				markdown += fmt.Sprintf("- **From**: `%s`\n", tx.From)
				markdown += fmt.Sprintf("- **To**: `%s`\n", tx.To)
				markdown += fmt.Sprintf("- **Amount**: `%s`\n", tx.Amount)
				markdown += fmt.Sprintf("- **Token**: `%s`\n", tx.Token)
				markdown += fmt.Sprintf("- **Type**: `%s`\n", tx.Type)
				markdown += fmt.Sprintf("- **Status**: `%s`\n", tx.Status)
				markdown += fmt.Sprintf("- **Confirmations**: `%d/%d`\n", tx.Confirmations, tx.RequiredConfirmations)
				markdown += fmt.Sprintf("- **Gas Fee**: `%s`\n", tx.GasFee)
				markdown += fmt.Sprintf("- **Priority**: `%s`\n", tx.Priority)
				markdown += fmt.Sprintf("- **Estimated Confirmation**: `%s`\n", tx.EstimatedConfirmationTime)
				markdown += fmt.Sprintf("- **Submitted**: `%s`\n", tx.SubmittedAt.Format("2006-01-02 15:04:05"))
				
				if tx.BlockNumber > 0 {
					markdown += fmt.Sprintf("- **Block**: `%d`\n", tx.BlockNumber)
				}
				
				if tx.Nonce > 0 {
					markdown += fmt.Sprintf("- **Nonce**: `%d`\n", tx.Nonce)
				}
				
				markdown += "\n"
			}
			
			// Add pagination info
			if offset > 0 || len(pendingTxs) == limit {
				markdown += "---\n"
				markdown += fmt.Sprintf("**Pagination**: Showing %d-%d", offset+1, offset+len(pendingTxs))
				if len(pendingTxs) == limit {
					markdown += " (more available)"
				}
				markdown += "\n"
			}
		}

		return mcp.NewToolResultText(markdown), nil
	}
}

