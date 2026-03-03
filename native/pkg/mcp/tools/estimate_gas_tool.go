package tools

import (
	"context"
	"fmt"

	"github.com/algonius/algonius-wallet/native/pkg/errors"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/toolutils"
	"github.com/algonius/algonius-wallet/native/pkg/wallet/chain"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// EstimateGasTool provides unified gas estimation across supported chains.
type EstimateGasTool struct {
	getChainInterface func(chainName string) (chain.IChain, error)
}

// NewEstimateGasTool creates a new EstimateGasTool.
func NewEstimateGasTool(factory *chain.ChainFactory) *EstimateGasTool {
	tool := &EstimateGasTool{
		getChainInterface: getChainInterface,
	}
	if factory != nil {
		tool.getChainInterface = factory.GetChain
	}
	return tool
}

// GetMeta returns MCP metadata for estimate_gas.
func (t *EstimateGasTool) GetMeta() mcp.Tool {
	return mcp.NewTool("estimate_gas",
		mcp.WithDescription("Estimate gas limit and gas price for a transaction"),
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
			mcp.Description("Amount to transfer"),
		),
		mcp.WithString("token",
			mcp.Description("Optional token symbol/contract (native token when omitted)"),
		),
	)
}

// GetHandler estimates gas for the requested transfer.
func (t *EstimateGasTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		chainName, err := req.RequireString("chain")
		if err != nil {
			return toolutils.FormatErrorResult(errors.MissingRequiredFieldError("chain")), nil
		}
		from, err := req.RequireString("from")
		if err != nil {
			return toolutils.FormatErrorResult(errors.MissingRequiredFieldError("from")), nil
		}
		to, err := req.RequireString("to")
		if err != nil {
			return toolutils.FormatErrorResult(errors.MissingRequiredFieldError("to")), nil
		}
		amount, err := req.RequireString("amount")
		if err != nil {
			return toolutils.FormatErrorResult(errors.MissingRequiredFieldError("amount")), nil
		}
		token := req.GetString("token", "")

		normalizedChain, err := toolutils.NormalizeChainName(chainName)
		if err != nil {
			if appErr, ok := err.(*errors.Error); ok {
				return toolutils.FormatErrorResult(appErr), nil
			}
			return toolutils.FormatErrorResult(errors.ValidationError("chain", err.Error())), nil
		}

		chainInterface, err := t.getChainInterface(normalizedChain)
		if err != nil {
			return toolutils.FormatErrorResult(errors.ValidationError("chain", fmt.Sprintf("unsupported chain: %s", normalizedChain))), nil
		}

		estimatedGas, err := toolutils.ExecuteWithRetry(ctx, toolutils.DefaultRetryPolicy, func(attemptCtx context.Context) (struct {
			gasLimit uint64
			gasPrice string
		}, error) {
			limit, price, estimateErr := chainInterface.EstimateGas(attemptCtx, from, to, amount, token)
			return struct {
				gasLimit uint64
				gasPrice string
			}{gasLimit: limit, gasPrice: price}, estimateErr
		})
		if err != nil {
			return toolutils.FormatErrorResult(toolutils.ClassifyError("estimate gas", err)), nil
		}

		markdown := fmt.Sprintf("### Gas Estimate\n\n- **Chain**: `%s`\n- **From**: `%s`\n- **To**: `%s`\n- **Amount**: `%s`\n- **Gas Limit**: `%d`\n- **Gas Price**: `%s`\n",
			normalizedChain, from, to, amount, estimatedGas.gasLimit, estimatedGas.gasPrice)
		if token != "" {
			markdown += fmt.Sprintf("- **Token**: `%s`\n", token)
		}

		return mcp.NewToolResultText(markdown), nil
	}
}
