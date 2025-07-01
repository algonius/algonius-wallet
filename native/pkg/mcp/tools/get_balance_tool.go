// Package tools provides MCP tool implementations for the Algonius Native Host.
package tools

import (
	"context"
	"fmt"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GetBalanceTool implements the MCP "get_balance" tool for querying wallet balance.
type GetBalanceTool struct {
	manager wallet.IWalletManager
}

// NewGetBalanceTool constructs a GetBalanceTool with the given wallet manager.
func NewGetBalanceTool(manager wallet.IWalletManager) *GetBalanceTool {
	return &GetBalanceTool{manager: manager}
}

// GetMeta returns the MCP tool definition for "get_balance" as per the documented API schema.
func (t *GetBalanceTool) GetMeta() mcp.Tool {
	return mcp.NewTool("get_balance",
		mcp.WithDescription("Query wallet balance"),
		mcp.WithString("address",
			mcp.Required(),
			mcp.Description("Wallet address"),
		),
		mcp.WithString("token",
			mcp.Required(),
			mcp.Description("Native token or token contract address"),
		),
	)
}

// GetHandler returns the handler function for the "get_balance" tool.
// The handler queries the balance for the specified address and token.
func (t *GetBalanceTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		address, err := req.RequireString("address")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid 'address' parameter"), nil
		}

		token, err := req.RequireString("token")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid 'token' parameter"), nil
		}

		balance, err := t.manager.GetBalance(ctx, address, token)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get balance: %v", err)), nil
		}

		markdown := "### Wallet Balance\n\n" +
			"- **Address**: `" + address + "`\n" +
			"- **Token**: `" + token + "`\n" +
			"- **Balance**: `" + balance + "`\n"
		return mcp.NewToolResultText(markdown), nil
	}
}
