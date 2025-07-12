// Package resources provides MCP resource implementations for the Algonius Native Host.
package resources

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WalletStatusResource implements the IResource interface for the "wallet_status" MCP resource.
// It returns the current wallet status including address, public key, ready state, and supported chains.
type WalletStatusResource struct {
	WalletManager wallet.IWalletManager
}

// NewWalletStatusResource creates a WalletStatusResource with the provided wallet manager.
func NewWalletStatusResource(walletManager wallet.IWalletManager) *WalletStatusResource {
	return &WalletStatusResource{
		WalletManager: walletManager,
	}
}

// titleCase converts the first character of a string to uppercase.
// This replaces the deprecated strings.Title function.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// GetMeta returns the MCP resource definition for wallet status.
func (r *WalletStatusResource) GetMeta() mcp.Resource {
	return mcp.NewResource(
		"wallet://status",
		"Wallet Status",
		mcp.WithResourceDescription("Query wallet status including address, public key, ready state, and supported chains"),
		mcp.WithMIMEType("text/markdown"),
	)
}

// GetHandler returns the handler function for the wallet status resource.
// The handler retrieves the current wallet status and returns it as Markdown resource contents.
func (r *WalletStatusResource) GetHandler() server.ResourceHandlerFunc {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Get wallet status from the wallet manager
		status, err := r.WalletManager.GetStatus(ctx)
		if err != nil {
			return nil, err
		}

		// Format the wallet status as AI-friendly Markdown
		markdown := r.formatWalletStatusMarkdown(status)

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "wallet://status",
				MIMEType: "text/markdown",
				Text:     markdown,
			},
		}, nil
	}
}

// formatWalletStatusMarkdown converts a WalletStatus struct to AI-friendly Markdown format.
func (r *WalletStatusResource) formatWalletStatusMarkdown(status *wallet.WalletStatus) string {
	var builder strings.Builder

	// Header
	builder.WriteString("# Wallet Status\n\n")

	// Overview section
	builder.WriteString("## Overview\n")
	
	// Ready status
	readyStatus := "Not Ready"
	if status.Ready {
		readyStatus = "Ready"
	}
	builder.WriteString(fmt.Sprintf("- **Status**: %s\n", readyStatus))

	// Address
	address := "Not created yet"
	if status.Address != "" {
		address = status.Address
	}
	builder.WriteString(fmt.Sprintf("- **Address**: %s\n", address))

	// Public Key
	publicKey := "Not created yet"
	if status.PublicKey != "" {
		// Truncate long public keys for better readability
		if len(status.PublicKey) > 20 {
			publicKey = status.PublicKey[:20] + "..."
		} else {
			publicKey = status.PublicKey
		}
	}
	builder.WriteString(fmt.Sprintf("- **Public Key**: %s\n", publicKey))

	// Last Used
	lastUsed := "Never"
	if status.LastUsed > 0 {
		lastUsed = time.Unix(status.LastUsed, 0).UTC().Format("2006-01-02 15:04:05 UTC")
	}
	builder.WriteString(fmt.Sprintf("- **Last Used**: %s\n\n", lastUsed))

	// Supported Chains section
	builder.WriteString("## Supported Chains\n")
	if len(status.Chains) > 0 {
		// Define display names for chains
		chainNames := map[string]string{
			"ethereum": "Ethereum (ETH)",
			"bsc":      "Binance Smart Chain (BSC)",
			"solana":   "Solana (SOL)",
		}

		// Sort chains for consistent output
		chains := []string{"ethereum", "bsc", "solana"}
		for _, chain := range chains {
			if supported, exists := status.Chains[chain]; exists {
				icon := "❌"
				if supported {
					icon = "✅"
				}
				displayName := chainNames[chain]
				if displayName == "" {
					displayName = titleCase(chain)
				}
				builder.WriteString(fmt.Sprintf("- %s %s\n", icon, displayName))
			}
		}
		
		// Add any additional chains not in the predefined list
		for chain, supported := range status.Chains {
			if chain != "ethereum" && chain != "bsc" && chain != "solana" {
				icon := "❌"
				if supported {
					icon = "✅"
				}
				builder.WriteString(fmt.Sprintf("- %s %s\n", icon, titleCase(chain)))
			}
		}
	} else {
		builder.WriteString("- No chains configured\n")
	}

	return builder.String()
}