// Package tools provides MCP tool implementations for the Algonius Native Host.
package tools

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/algonius/algonius-wallet/native/pkg/wallet/dex"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SwapTokensTool implements the MCP "swap_tokens" tool for token swapping through DEX protocols.
type SwapTokensTool struct {
	manager    wallet.IWalletManager
	dexFactory dex.IDEXFactory
}

// NewSwapTokensTool constructs a SwapTokensTool with the given wallet manager.
func NewSwapTokensTool(manager wallet.IWalletManager) *SwapTokensTool {
	return &SwapTokensTool{
		manager:    manager,
		dexFactory: dex.NewDEXFactory(),
	}
}

// GetMeta returns the MCP tool definition for "swap_tokens" as per the documented API schema.
func (t *SwapTokensTool) GetMeta() mcp.Tool {
	return mcp.NewTool("swap_tokens",
		mcp.WithDescription("Swap tokens through DEX protocols"),
		mcp.WithString("chain",
			mcp.Required(),
			mcp.Description("Chain identifier (ethereum, bsc)"),
		),
		mcp.WithString("token_in",
			mcp.Required(),
			mcp.Description("Input token contract address"),
		),
		mcp.WithString("token_out",
			mcp.Required(),
			mcp.Description("Output token contract address"),
		),
		mcp.WithString("amount_in",
			mcp.Description("Input amount (for exact input swaps)"),
		),
		mcp.WithString("amount_out",
			mcp.Description("Output amount (for exact output swaps)"),
		),
		mcp.WithNumber("slippage_tolerance",
			mcp.Description("Maximum slippage tolerance in percent (default: 0.5)"),
		),
		mcp.WithNumber("deadline",
			mcp.Description("Transaction deadline in seconds (default: 300)"),
		),
		mcp.WithString("dex",
			mcp.Description("DEX protocol to use (default: uniswap)"),
		),
		mcp.WithString("from",
			mcp.Required(),
			mcp.Description("Sender/recipient address"),
		),
	)
}

// GetHandler returns the handler function for the "swap_tokens" tool.
func (t *SwapTokensTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract required parameters
		chain, err := req.RequireString("chain")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid 'chain' parameter"), nil
		}

		tokenIn, err := req.RequireString("token_in")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid 'token_in' parameter"), nil
		}

		tokenOut, err := req.RequireString("token_out")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid 'token_out' parameter"), nil
		}

		from, err := req.RequireString("from")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid 'from' parameter"), nil
		}

		// Extract optional parameters
		amountInStr := req.GetString("amount_in", "")
		amountOutStr := req.GetString("amount_out", "")
		slippageTolerance := req.GetFloat("slippage_tolerance", 0.5)
		deadline := req.GetFloat("deadline", 300)
		dexProtocol := req.GetString("dex", "uniswap")

		// Validate that either amount_in or amount_out is provided
		if amountInStr == "" && amountOutStr == "" {
			return mcp.NewToolResultError("either 'amount_in' or 'amount_out' must be specified"), nil
		}

		if amountInStr != "" && amountOutStr != "" {
			return mcp.NewToolResultError("only one of 'amount_in' or 'amount_out' should be specified"), nil
		}

		// Validate chain support
		if chain != "ethereum" && chain != "bsc" && chain != "eth" {
			return mcp.NewToolResultError(fmt.Sprintf("unsupported chain: %s. Supported chains: ethereum, bsc", chain)), nil
		}

		// Parse amounts
		var amountIn, amountOut *big.Int
		if amountInStr != "" {
			amountIn, err = t.parseAmount(amountInStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid amount_in: %v", err)), nil
			}
		}

		if amountOutStr != "" {
			amountOut, err = t.parseAmount(amountOutStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid amount_out: %v", err)), nil
			}
		}

		// Validate slippage tolerance
		if slippageTolerance < 0.1 || slippageTolerance > 50.0 {
			return mcp.NewToolResultError(fmt.Sprintf("slippage_tolerance must be between 0.1 and 50.0, got %.2f", slippageTolerance)), nil
		}

		// Calculate deadline timestamp
		deadlineTimestamp := uint64(time.Now().Unix() + int64(deadline))

		// Create DEX instance
		dexInstance, err := t.dexFactory.CreateDEX(dexProtocol, chain)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create DEX instance: %v", err)), nil
		}

		// Create swap parameters
		swapParams := &dex.SwapParams{
			TokenIn:           tokenIn,
			TokenOut:          tokenOut,
			AmountIn:          amountIn,
			AmountOut:         amountOut,
			SlippageTolerance: slippageTolerance,
			Deadline:          deadlineTimestamp,
			Recipient:         from, // Use sender as recipient for simplicity
			From:              from,
			PrivateKey:        "0x0000000000000000000000000000000000000000000000000000000000000001", // Mock private key
		}

		// Get quote first for validation
		_, err = dexInstance.GetQuote(ctx, swapParams)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get swap quote: %v", err)), nil
		}

		// Perform the swap
		var result *dex.SwapResult
		if amountIn != nil {
			// Exact input swap
			result, err = dexInstance.SwapExactTokensForTokens(ctx, swapParams)
		} else {
			// Exact output swap
			result, err = dexInstance.SwapTokensForExactTokens(ctx, swapParams)
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to execute swap: %v", err)), nil
		}

		// Format success response
		markdown := "### Token Swap Executed\n\n" +
			"- **Chain**: `" + chain + "`\n" +
			"- **DEX**: `" + dexInstance.GetProtocolName() + "`\n" +
			"- **From**: `" + from + "`\n" +
			"- **Token In**: `" + tokenIn + "`\n" +
			"- **Token Out**: `" + tokenOut + "`\n" +
			"- **Amount In**: `" + result.AmountIn.String() + "`\n" +
			"- **Amount Out**: `" + result.AmountOut.String() + "`\n" +
			fmt.Sprintf("- **Price Impact**: `%.2f%%`\n", result.PriceImpact) +
			fmt.Sprintf("- **Slippage Tolerance**: `%.2f%%`\n", slippageTolerance) +
			"- **Route**: `" + strings.Join(result.Route, " → ") + "`\n" +
			"- **Transaction Hash**: `" + result.TransactionHash + "`\n" +
			fmt.Sprintf("- **Gas Used**: `%d`\n", result.GasUsed) +
			"- **Status**: `completed`\n"

		return mcp.NewToolResultText(markdown), nil
	}
}

// parseAmount parses amount string to big.Int, handling both integer and decimal representations
func (t *SwapTokensTool) parseAmount(amountStr string) (*big.Int, error) {
	// Remove any whitespace
	amountStr = strings.TrimSpace(amountStr)

	// Try parsing as float first to handle decimal amounts
	if strings.Contains(amountStr, ".") {
		// Parse as float and convert to wei (assuming 18 decimals for simplicity)
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid decimal amount: %s", amountStr)
		}

		if amount <= 0 {
			return nil, fmt.Errorf("amount must be positive: %f", amount)
		}

		// Convert to wei (multiply by 10^18)
		weiAmount := amount * 1e18
		return big.NewInt(int64(weiAmount)), nil
	}

	// Parse as integer
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount format: %s", amountStr)
	}

	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("amount must be positive: %s", amountStr)
	}

	return amount, nil
}