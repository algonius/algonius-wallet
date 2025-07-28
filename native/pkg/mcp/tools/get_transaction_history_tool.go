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

// GetTransactionHistoryTool implements the MCP "get_transaction_history" tool for querying transaction history.
type GetTransactionHistoryTool struct {
	manager wallet.IWalletManager
}

// NewGetTransactionHistoryTool constructs a GetTransactionHistoryTool with the given wallet manager.
func NewGetTransactionHistoryTool(manager wallet.IWalletManager) *GetTransactionHistoryTool {
	return &GetTransactionHistoryTool{manager: manager}
}

// GetMeta returns the MCP tool definition for "get_transaction_history" as per the documented API schema.
func (t *GetTransactionHistoryTool) GetMeta() mcp.Tool {
	return mcp.NewTool("get_transaction_history",
		mcp.WithDescription("Query transaction history for a wallet address"),
		mcp.WithString("address",
			mcp.Description("Wallet address to query transaction history for"),
			mcp.Required(),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of transactions to return (default: 10, max: 100)"),
		),
		mcp.WithNumber("from_block",
			mcp.Description("Optional starting block number"),
		),
		mcp.WithNumber("to_block",
			mcp.Description("Optional ending block number"),
		),
	)
}

// GetHandler returns the handler function for the "get_transaction_history" tool.
// The handler queries transaction history with optional filtering and pagination.
func (t *GetTransactionHistoryTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse required parameters
		address := req.GetString("address", "")
		if address == "" {
			toolErr := errors.ValidationError("get transaction history", "address is required")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Parse pagination parameters
		limit := int(req.GetFloat("limit", 10))
		if limit <= 0 {
			limit = 10 // Default limit
		}
		if limit > 100 {
			limit = 100 // Maximum limit
		}

		// Parse optional block range parameters
		var fromBlock, toBlock *uint64
		if fromBlockFloat := req.GetFloat("from_block", -1); fromBlockFloat >= 0 {
			blockNum := uint64(fromBlockFloat)
			fromBlock = &blockNum
		}
		if toBlockFloat := req.GetFloat("to_block", -1); toBlockFloat >= 0 {
			blockNum := uint64(toBlockFloat)
			toBlock = &blockNum
		}

		// Get transaction history from wallet manager
		transactions, err := t.manager.GetTransactionHistory(ctx, address, fromBlock, toBlock, limit, 0)
		if err != nil {
			toolErr := errors.InternalError("get transaction history", err)
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Format response as markdown
		markdown := "### Transaction History\n\n"

		if len(transactions) == 0 {
			markdown += "No transactions found for the specified address.\n"
		} else {
			markdown += fmt.Sprintf("Found %d transactions for address `%s`:\n\n", len(transactions), address)

			for i, tx := range transactions {
				markdown += fmt.Sprintf("#### Transaction %d\n", i+1)
				markdown += fmt.Sprintf("- **Hash**: `%s`\n", tx.Hash)
				markdown += fmt.Sprintf("- **Chain**: `%s`\n", tx.Chain)
				markdown += fmt.Sprintf("- **Block**: `%d`\n", tx.BlockNumber)
				markdown += fmt.Sprintf("- **From**: `%s`\n", tx.From)
				markdown += fmt.Sprintf("- **To**: `%s`\n", tx.To)
				markdown += fmt.Sprintf("- **Value**: `%s`\n", tx.Value)

				if tx.TokenSymbol != "" && tx.TokenSymbol != "ETH" && tx.TokenSymbol != "BNB" {
					markdown += fmt.Sprintf("- **Token**: `%s`\n", tx.TokenSymbol)
				}

				markdown += fmt.Sprintf("- **Type**: `%s`\n", tx.Type)
				markdown += fmt.Sprintf("- **Status**: `%s`\n", tx.Status)
				markdown += fmt.Sprintf("- **Fee**: `%s`\n", tx.TransactionFee)
				markdown += fmt.Sprintf("- **Confirmations**: `%d`\n", tx.Confirmations)
				markdown += fmt.Sprintf("- **Timestamp**: `%s`\n", tx.Timestamp.Format("2006-01-02 15:04:05"))

				if tx.GasUsed != "" {
					markdown += fmt.Sprintf("- **Gas Used**: `%s`\n", tx.GasUsed)
				}

				if tx.ContractAddress != "" {
					markdown += fmt.Sprintf("- **Contract**: `%s`\n", tx.ContractAddress)
					if tx.MethodName != "" {
						markdown += fmt.Sprintf("- **Method**: `%s`\n", tx.MethodName)
					}
				}

				// Add token transfers if present
				if len(tx.TokenTransfers) > 0 {
					markdown += "- **Token Transfers**:\n"
					for j, transfer := range tx.TokenTransfers {
						markdown += fmt.Sprintf("  - Transfer %d: `%s` %s from `%s` to `%s`\n", 
							j+1, transfer.Value, transfer.TokenSymbol, transfer.From, transfer.To)
					}
				}

				markdown += "\n"
			}

			// Add block range info if specified
			if fromBlock != nil || toBlock != nil {
				markdown += "---\n"
				markdown += "**Block Range**: "
				if fromBlock != nil {
					markdown += fmt.Sprintf("from %d ", *fromBlock)
				}
				if toBlock != nil {
					markdown += fmt.Sprintf("to %d ", *toBlock)
				}
				markdown += "\n"
			}

			// Add pagination info
			if len(transactions) == limit {
				markdown += "---\n"
				markdown += fmt.Sprintf("**Note**: Showing first %d transactions (more may be available)\n", limit)
			}
		}

		return mcp.NewToolResultText(markdown), nil
	}
}