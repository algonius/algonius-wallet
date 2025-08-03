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
	"go.uber.org/zap"
)

// SignMessageTool implements the MCP "sign_message" tool for signing messages with wallet private keys.
type SignMessageTool struct {
	manager wallet.IWalletManager
	logger  *zap.Logger
}

// NewSignMessageTool constructs a SignMessageTool with the given wallet manager.
func NewSignMessageTool(manager wallet.IWalletManager, logger *zap.Logger) *SignMessageTool {
	if logger == nil {
		logger = zap.NewNop() // Use no-op logger if none provided
	}
	return &SignMessageTool{
		manager: manager,
		logger:  logger,
	}
}

// GetMeta returns the MCP tool definition for "sign_message" as per the documented API schema.
func (t *SignMessageTool) GetMeta() mcp.Tool {
	return mcp.NewTool("sign_message",
		mcp.WithDescription("Sign a text message or raw bytes with a wallet's private key"),
		mcp.WithString("address",
			mcp.Required(),
			mcp.Description("The wallet address to sign with"),
		),
		mcp.WithString("message",
			mcp.Required(),
			mcp.Description("The message to sign. For Solana raw bytes, prefix with '__SOLANA_RAW_BYTES__:'"),
		),
	)
}

// GetHandler returns the handler function for the "sign_message" tool.
// The handler signs messages using the wallet's private key for the specified address.
func (t *SignMessageTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract and validate address parameter
		address, err := req.RequireString("address")
		if err != nil {
			toolErr := errors.MissingRequiredFieldError("address")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Extract and validate message parameter
		message, err := req.RequireString("message")
		if err != nil {
			toolErr := errors.MissingRequiredFieldError("message")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Validate address format
		if address == "" {
			toolErr := errors.ValidationError("address", "address cannot be empty")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Validate message format
		if message == "" {
			toolErr := errors.ValidationError("message", "message cannot be empty")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		t.logger.Debug("Signing message with wallet address",
			zap.String("address", address),
			zap.Int("message_length", len(message)),
			zap.Bool("is_solana_raw_bytes", len(message) > 20 && message[:20] == "__SOLANA_RAW_BYTES__"))

		// Sign the message using the wallet manager
		signature, err := t.manager.SignMessage(ctx, address, message)
		if err != nil {
			t.logger.Error("Failed to sign message",
				zap.String("address", address),
				zap.Error(err))
			toolErr := errors.InternalError("sign message", err)
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Determine chain type for response formatting
		var chainType string
		if len(message) > 20 && message[:20] == "__SOLANA_RAW_BYTES__" {
			chainType = "solana"
		} else if len(address) > 2 && address[:2] == "0x" {
			chainType = "ethereum"
		} else {
			chainType = "solana" // Default to Solana for base58 addresses
		}

		t.logger.Info("Message signed successfully",
			zap.String("address", address),
			zap.String("chain_type", chainType),
			zap.String("signature_length", fmt.Sprintf("%d", len(signature))))

		// Format response as markdown
		markdown := fmt.Sprintf("### Message Signed Successfully ✅\n\n"+
			"- **Address**: `%s`\n"+
			"- **Chain**: `%s`\n"+
			"- **Message Length**: `%d characters`\n"+
			"- **Signature**: `%s`\n"+
			"- **Action**: Message has been cryptographically signed with the wallet's private key\n",
			address, chainType, len(message), signature)

		return mcp.NewToolResultText(markdown), nil
	}
}