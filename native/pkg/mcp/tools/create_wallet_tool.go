// Package tools provides MCP tool implementations for the Algonius Native Host.
package tools

import (
	"context"

	"github.com/algonius/algonius-wallet/native/pkg/errors"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/toolutils"
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
			toolErr := errors.MissingRequiredFieldError("chain")
			return toolutils.FormatErrorResult(toolErr), nil
		}
		// MCP tools don't have access to user passwords, so we use a default
		// This is primarily for AI agent interactions, not end-user wallet creation
		defaultPassword := "temp-mcp-password-123"
		address, publicKey, mnemonic, err := t.manager.CreateWallet(ctx, chain, defaultPassword)
		if err != nil {
			toolErr := errors.InternalError("create wallet", err)
			return toolutils.FormatErrorResult(toolErr), nil
		}
		// Note: mnemonic is not included in tool output for security reasons
		_ = mnemonic // Acknowledge that we received the mnemonic but don't expose it to AI
		markdown := "### Wallet Created\n\n" +
			"- **Address**: `" + address + "`\n" +
			"- **Public Key**: `" + publicKey + "`\n"
		return mcp.NewToolResultText(markdown), nil
	}
}
