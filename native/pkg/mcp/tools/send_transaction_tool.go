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
			mcp.Description("Chain identifier: ethereum|eth, bsc|binance, solana|sol"),
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
			toolErr := errors.MissingRequiredFieldError("chain")
			return toolutils.FormatErrorResult(toolErr), nil
		}
		normalizedChain, err := toolutils.NormalizeChainName(chain)
		if err != nil {
			if appErr, ok := err.(*errors.Error); ok {
				return toolutils.FormatErrorResult(appErr), nil
			}
			return toolutils.FormatErrorResult(errors.ValidationError("chain", err.Error())), nil
		}

		from, err := req.RequireString("from")
		if err != nil {
			toolErr := errors.MissingRequiredFieldError("from")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		to, err := req.RequireString("to")
		if err != nil {
			toolErr := errors.MissingRequiredFieldError("to")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		amount, err := req.RequireString("amount")
		if err != nil {
			toolErr := errors.MissingRequiredFieldError("amount")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Extract optional parameters
		token := req.GetString("token", "")
		gasLimit := req.GetFloat("gas_limit", 0)
		gasPrice := req.GetString("gas_price", "")

		// Perform gas estimation if not provided
		var finalGasLimit float64 = gasLimit
		var finalGasPrice string = gasPrice

		if gasLimit == 0 || gasPrice == "" {
			estimatedGas, err := toolutils.ExecuteWithRetry(ctx, toolutils.DefaultRetryPolicy, func(attemptCtx context.Context) (struct {
				gasLimit uint64
				gasPrice string
			}, error) {
				limit, price, estimateErr := t.manager.EstimateGas(attemptCtx, normalizedChain, from, to, amount, token)
				return struct {
					gasLimit uint64
					gasPrice string
				}{gasLimit: limit, gasPrice: price}, estimateErr
			})
			if err != nil {
				toolErr := toolutils.ClassifyError("gas estimation", err)
				return toolutils.FormatErrorResult(toolErr), nil
			}

			if gasLimit == 0 {
				finalGasLimit = float64(estimatedGas.gasLimit)
			}
			if gasPrice == "" {
				finalGasPrice = estimatedGas.gasPrice
			}
		}

		// Send the transaction
		txHash, err := toolutils.ExecuteWithRetry(ctx, toolutils.DefaultRetryPolicy, func(attemptCtx context.Context) (string, error) {
			return t.manager.SendTransaction(attemptCtx, normalizedChain, from, to, amount, token)
		})
		if err != nil {
			toolErr := toolutils.ClassifyError("send transaction", err)
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Format success response
		markdown := "### Transaction Sent\n\n" +
			"- **Chain**: `" + normalizedChain + "`\n" +
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
