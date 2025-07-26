// Package tools provides MCP tool implementations for the Algonius Native Host.
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/algonius/algonius-wallet/native/pkg/simulation"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/algonius/algonius-wallet/native/pkg/wallet/chain"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SimulateTransactionTool implements the MCP "simulate_transaction" tool for transaction simulation.
type SimulateTransactionTool struct {
	manager      wallet.IWalletManager
	simulator    *simulation.TransactionSimulator
	chainFactory *chain.ChainFactory
}

// NewSimulateTransactionTool constructs a SimulateTransactionTool with the given dependencies.
func NewSimulateTransactionTool(manager wallet.IWalletManager, chainFactory *chain.ChainFactory) *SimulateTransactionTool {
	return &SimulateTransactionTool{
		manager:      manager,
		simulator:    simulation.NewTransactionSimulator(chainFactory),
		chainFactory: chainFactory,
	}
}

// GetMeta returns the MCP tool definition for "simulate_transaction" as per the documented API schema.
func (t *SimulateTransactionTool) GetMeta() mcp.Tool {
	return mcp.NewTool("simulate_transaction",
		mcp.WithDescription("Simulate a blockchain transaction without executing it"),
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
	)
}

// GetHandler returns the handler function for the "simulate_transaction" tool.
// The handler simulates a transaction on the specified blockchain without executing it.
func (t *SimulateTransactionTool) GetHandler() server.ToolHandlerFunc {
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

		// Validate chain support
		if chain != "ethereum" && chain != "bsc" && chain != "ETH" {
			return mcp.NewToolResultError(
				"unsupported chain: " + chain + ". Supported chains: ethereum, bsc"), nil
		}

		// Simulate the transaction
		result, err := t.simulator.SimulateTransaction(ctx, chain, from, to, amount, token)
		if err != nil {
			return mcp.NewToolResultError("failed to simulate transaction: " + err.Error()), nil
		}

		// Convert result to JSON
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("failed to marshal simulation result: " + err.Error()), nil
		}

		// Format success response
		markdown := "### Transaction Simulation\n\n" +
			"- **Chain**: `" + chain + "`\n" +
			"- **From**: `" + from + "`\n" +
			"- **To**: `" + to + "`\n" +
			"- **Amount**: `" + amount + "`\n"

		if token != "" {
			markdown += "- **Token**: `" + token + "`\n"
		}

		if result.Success {
			markdown += "- **Success**: `true`\n" +
				"- **Gas Used**: `" + fmt.Sprintf("%d", result.GasUsed) + "`\n" +
				"- **Gas Price**: `" + result.GasPrice + " gwei`\n" +
				"- **Total Cost**: `" + result.TotalCost + "`\n" +
				"- **Balance Change**: `" + result.BalanceChange + "`\n"

			if len(result.Warnings) > 0 {
				markdown += "- **Warnings**:\n"
				for _, warning := range result.Warnings {
					markdown += "  - " + warning + "\n"
				}
			}
		} else {
			markdown += "- **Success**: `false`\n"
			if len(result.Errors) > 0 {
				markdown += "- **Errors**:\n"
				for _, error := range result.Errors {
					markdown += "  - " + error + "\n"
				}
			}
		}

		// Create a tool result with both markdown text and JSON data
		toolResult := mcp.NewToolResultText(markdown)
		// Add metadata with the raw JSON result
		if toolResult.Meta == nil {
			toolResult.Meta = make(map[string]any)
		}
		toolResult.Meta["json_result"] = string(resultJSON)
		return toolResult, nil
	}
}