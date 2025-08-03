// Package tools provides MCP tool implementations for the Algonius Native Host.
package tools

import (
	"context"

	"github.com/algonius/algonius-wallet/native/pkg/errors"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/toolutils"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/algonius/algonius-wallet/native/pkg/wallet/chain"
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
	description := "Query wallet balance for native tokens and contracts. " +
		"Supported native tokens: ETH (Ethereum), BNB (BSC), SOL (Solana). " +
		"Also supports ERC-20/BEP-20 contract addresses."
	
	tokenDescription := "Token identifier or contract address. " +
		"Native tokens: ETH, ETHER, BNB, BINANCE, SOL, SOLANA. " +
		"Contract addresses: 0x... (Ethereum/BSC) or base58 (Solana)"
	
	return mcp.NewTool("get_balance",
		mcp.WithDescription(description),
		mcp.WithString("address",
			mcp.Required(),
			mcp.Description("Wallet address (0x... for ETH/BSC, base58 for Solana)"),
		),
		mcp.WithString("token",
			mcp.Required(),
			mcp.Description(tokenDescription),
		),
	)
}

// GetHandler returns the handler function for the "get_balance" tool.
// The handler queries the balance for the specified address and token.
func (t *GetBalanceTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		address, err := req.RequireString("address")
		if err != nil {
			toolErr := errors.MissingRequiredFieldError("address")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		token, err := req.RequireString("token")
		if err != nil {
			toolErr := errors.MissingRequiredFieldError("token")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Validate token using centralized mapping for better error messages
		if _, err := chain.DefaultTokenMapping.GetTokenInfo(token); err != nil {
			// Check if it might be a contract address
			if !((len(token) == 42 && token[:2] == "0x") || (len(token) >= 32 && len(token) <= 44)) {
				// Provide helpful suggestion for unsupported tokens
				toolErr := errors.New(errors.ErrTokenNotSupported, "Token identifier not recognized").
					WithDetails(err.Error()).
					WithSuggestion("Supported tokens: ETH, ETHER, BNB, BINANCE, SOL, SOLANA. For contract tokens, use full address.")
				return toolutils.FormatErrorResult(toolErr), nil
			}
		}

		balance, err := t.manager.GetBalance(ctx, address, token)
		if err != nil {
			toolErr := errors.InternalError("get balance", err)
			return toolutils.FormatErrorResult(toolErr), nil
		}

		markdown := "### Wallet Balance\n\n" +
			"- **Address**: `" + address + "`\n" +
			"- **Token**: `" + token + "`\n" +
			"- **Balance**: `" + balance + "`\n"
		return mcp.NewToolResultText(markdown), nil
	}
}

