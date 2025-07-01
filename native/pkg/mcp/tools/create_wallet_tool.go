// Package tools provides MCP tool implementations for the Algonius Native Host.
package tools

import (
	"context"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// CreateWalletTool implements the MCP "create_wallet" tool for generating a new wallet.
type CreateWalletTool struct {
	manager wallet.IWalletManager
}

// NewCreateWalletTool constructs a CreateWalletTool with the given wallet manager.
func NewCreateWalletTool(manager wallet.IWalletManager) *CreateWalletTool {
	return &CreateWalletTool{manager: manager}
}

// GetMeta returns the MCP tool definition for "create_wallet" as per the documented API schema.
func (t *CreateWalletTool) GetMeta() mcp.Tool {
	return mcp.NewTool("create_wallet",
		mcp.WithDescription("Create a new wallet (generate private key locally)"),
		mcp.WithString("chain",
			mcp.Required(),
			mcp.Description("Chain identifier, e.g. ETH"),
		),
	)
}

// GetHandler returns the handler function for the "create_wallet" tool.
// The handler creates a new wallet for the specified chain and returns its address and public key.
func (t *CreateWalletTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		chain, err := req.RequireString("chain")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid 'chain' parameter"), nil
		}
		address, publicKey, err := t.manager.CreateWallet(ctx, chain)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		markdown := "### Wallet Created\n\n" +
			"- **Address**: `" + address + "`\n" +
			"- **Public Key**: `" + publicKey + "`\n"
		return mcp.NewToolResultText(markdown), nil
	}
}
