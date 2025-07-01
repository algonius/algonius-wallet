// Package tools provides MCP tool implementations for the Algonius Native Host.
package tools

import (
	"context"
	"fmt"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SendTransactionTool implements the MCP "send_transaction" tool for sending blockchain transactions.
type SendTransactionTool struct {
	manager wallet.IWalletManager
}

// NewSendTransactionTool constructs a SendTransactionTool with the given wallet manager.
func NewSendTransactionTool(manager wallet.IWalletManager) *SendTransactionTool {
	return &SendTransactionTool{manager: manager}
}

// GetMeta returns the MCP tool definition for "send_transaction" as per the documented API schema.
func (t *SendTransactionTool) GetMeta() mcp.Tool {
	return mcp.NewTool("send_transaction",
		mcp.WithDescription("Send a blockchain transaction"),
		mcp.WithString("chain",
			mcp.Required(),
			mcp.Description("Chain identifier (ethereum, bsc)"),
		),
		mcp.WithString("from",
			mcp.Required(),
			mcp.Description("Sender address"),
		),
		mcp.WithString("to",
			mcp.Required(),
			mcp.Description("Recipient address"),
		),
		mcp.WithString("amount",
			mcp.Required(),
			mcp.Description("Amount to send"),
		),
		mcp.WithString("token",
			mcp.Description("Token contract address (optional, native token if not provided)"),
		),
		mcp.WithNumber("gas_limit",
			mcp.Description("Gas limit (optional)"),
		),
		mcp.WithString("gas_price",
			mcp.Description("Gas price in gwei (optional)"),
		),
	)
}

// GetHandler returns the handler function for the "send_transaction" tool.
// The handler sends a transaction on the specified blockchain.
func (t *SendTransactionTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract required parameters
		chain, err := req.RequireString("chain")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid 'chain' parameter"), nil
		}

		from, err := req.RequireString("from")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid 'from' parameter"), nil
		}

		to, err := req.RequireString("to")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid 'to' parameter"), nil
		}

		amount, err := req.RequireString("amount")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid 'amount' parameter"), nil
		}

		// Extract optional parameters
		token := req.GetString("token", "")
		gasLimit := req.GetFloat("gas_limit", 0)
		gasPrice := req.GetString("gas_price", "")

		// Validate chain support
		if chain != "ethereum" && chain != "bsc" && chain != "ETH" {
			return mcp.NewToolResultError(fmt.Sprintf("unsupported chain: %s. Supported chains: ethereum, bsc", chain)), nil
		}

		// Perform gas estimation if not provided
		var finalGasLimit float64 = gasLimit
		var finalGasPrice string = gasPrice

		if gasLimit == 0 || gasPrice == "" {
			estimatedGasLimit, estimatedGasPrice, err := t.manager.EstimateGas(ctx, chain, from, to, amount, token)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to estimate gas: %v", err)), nil
			}

			if gasLimit == 0 {
				finalGasLimit = float64(estimatedGasLimit)
			}
			if gasPrice == "" {
				finalGasPrice = estimatedGasPrice
			}
		}

		// Send the transaction
		txHash, err := t.manager.SendTransaction(ctx, chain, from, to, amount, token)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to send transaction: %v", err)), nil
		}

		// Format success response
		markdown := "### Transaction Sent\n\n" +
			"- **Chain**: `" + chain + "`\n" +
			"- **From**: `" + from + "`\n" +
			"- **To**: `" + to + "`\n" +
			"- **Amount**: `" + amount + "`\n"

		if token != "" {
			markdown += "- **Token**: `" + token + "`\n"
		}

		if finalGasLimit > 0 {
			markdown += fmt.Sprintf("- **Gas Limit**: `%.0f`\n", finalGasLimit)
		}

		if finalGasPrice != "" {
			markdown += "- **Gas Price**: `" + finalGasPrice + " gwei`\n"
		}

		markdown += "- **Transaction Hash**: `" + txHash + "`\n" +
			"- **Status**: `pending`\n"

		return mcp.NewToolResultText(markdown), nil
	}
}