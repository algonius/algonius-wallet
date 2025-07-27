// Package tools provides MCP tool implementations for the Algonius Native Host.
package tools

import (
	"context"
	"fmt"

	"github.com/algonius/algonius-wallet/native/pkg/dex"
	"github.com/algonius/algonius-wallet/native/pkg/dex/providers"
	"github.com/algonius/algonius-wallet/native/pkg/errors"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/toolutils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// SwapTokensToolNew implements the swap_tokens tool using the new DEX architecture
type SwapTokensToolNew struct {
	dexAggregator dex.IDEXAggregator
	logger        *zap.Logger
}

// NewSwapTokensToolNew creates a new SwapTokensToolNew instance
func NewSwapTokensToolNew(logger *zap.Logger) *SwapTokensToolNew {
	// Initialize DEX aggregator with direct provider for now
	aggregator := dex.NewDEXAggregator(logger)
	directProvider := providers.NewDirectProvider(logger)
	aggregator.RegisterProvider(directProvider)

	return &SwapTokensToolNew{
		dexAggregator: aggregator,
		logger:        logger,
	}
}

// Definition returns the tool definition for MCP
func (t *SwapTokensToolNew) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "swap_tokens_new",
		Description: "Execute token swaps using the new DEX aggregator system. Supports multiple DEX providers and chains.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"chain": map[string]interface{}{
					"type":        "string",
					"description": "Blockchain to use for the swap",
					"enum":        []string{"ethereum", "bsc", "solana"},
				},
				"from_token": map[string]interface{}{
					"type":        "string", 
					"description": "Token to swap from (address or symbol)",
				},
				"to_token": map[string]interface{}{
					"type":        "string",
					"description": "Token to swap to (address or symbol)", 
				},
				"amount": map[string]interface{}{
					"type":        "string",
					"description": "Amount to swap",
				},
				"from_address": map[string]interface{}{
					"type":        "string",
					"description": "Address initiating the swap",
				},
				"slippage": map[string]interface{}{
					"type":        "number",
					"description": "Maximum acceptable slippage (e.g., 0.005 for 0.5%)",
					"default":     0.005,
				},
			},
			Required: []string{"chain", "from_token", "to_token", "amount", "from_address"},
		},
	}
}

// Execute performs the token swap
func (t *SwapTokensToolNew) Execute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		toolErr := errors.ValidationError("arguments", "invalid arguments format")
		return toolutils.FormatErrorResult(toolErr), nil
	}
	
	chain, _ := arguments["chain"].(string)
	fromToken, _ := arguments["from_token"].(string)
	toToken, _ := arguments["to_token"].(string)
	amount, _ := arguments["amount"].(string)
	fromAddress, _ := arguments["from_address"].(string)
	slippage, _ := arguments["slippage"].(float64)

	// Set default slippage if not provided
	if slippage == 0 {
		slippage = 0.005 // 0.5%
	}

	// Validate parameters
	if chain == "" {
		toolErr := errors.ValidationError("chain", "chain parameter is required")
		return toolutils.FormatErrorResult(toolErr), nil
	}
	if fromToken == "" {
		toolErr := errors.ValidationError("from_token", "from_token parameter is required")
		return toolutils.FormatErrorResult(toolErr), nil
	}
	if toToken == "" {
		toolErr := errors.ValidationError("to_token", "to_token parameter is required")
		return toolutils.FormatErrorResult(toolErr), nil
	}
	if amount == "" {
		toolErr := errors.ValidationError("amount", "amount parameter is required")
		return toolutils.FormatErrorResult(toolErr), nil
	}
	if fromAddress == "" {
		toolErr := errors.ValidationError("from_address", "from_address parameter is required")
		return toolutils.FormatErrorResult(toolErr), nil
	}

	// Map chain name to chain ID
	chainID := t.mapChainNameToID(chain)
	if chainID == "" {
		toolErr := errors.ValidationError("chain", fmt.Sprintf("unsupported chain: %s", chain))
		return toolutils.FormatErrorResult(toolErr), nil
	}

	// Create swap parameters
	swapParams := dex.SwapParams{
		FromToken:   fromToken,
		ToToken:     toToken,
		Amount:      amount,
		Slippage:    slippage,
		FromAddress: fromAddress,
		ToAddress:   fromAddress, // Use same address as recipient
		ChainID:     chainID,
		PrivateKey:  "0x0000000000000000000000000000000000000000000000000000000000000001", // Mock private key for demo
	}

	// Get quote first
	quote, err := t.dexAggregator.GetBestQuote(ctx, swapParams)
	if err != nil {
		toolErr := errors.InternalError("get swap quote", err)
		return toolutils.FormatErrorResult(toolErr), nil
	}

	// Execute the swap
	result, err := t.dexAggregator.ExecuteSwapWithProvider(ctx, quote.Provider, swapParams)
	if err != nil {
		toolErr := errors.InternalError("execute swap", err)
		return toolutils.FormatErrorResult(toolErr), nil
	}

	// Format success response
	markdown := fmt.Sprintf(`### Token Swap Executed

- **Chain**: %s
- **Provider**: %s
- **From**: %s
- **Token In**: %s
- **Token Out**: %s
- **Amount In**: %s
- **Amount Out**: %s
- **Slippage**: %.2f%%
- **Estimated Fee**: %s
- **Transaction Hash**: %s
- **Status**: %s

The swap has been executed successfully!`, 
		chain,
		quote.Provider,
		fromAddress,
		fromToken,
		toToken,
		result.FromAmount,
		result.ToAmount,
		slippage*100,
		result.ActualFee,
		result.TxHash,
		result.Status)

	return mcp.NewToolResultText(markdown), nil
}

// mapChainNameToID maps human-readable chain names to chain IDs
func (t *SwapTokensToolNew) mapChainNameToID(chainName string) string {
	switch chainName {
	case "ethereum", "eth":
		return "1"
	case "bsc", "binance":
		return "56"
	case "solana", "sol":
		return "501"
	default:
		return ""
	}
}

// Register registers the tool with the MCP server
func (t *SwapTokensToolNew) Register(srv *server.MCPServer) {
	srv.AddTool(t.Definition(), t.Execute)
}