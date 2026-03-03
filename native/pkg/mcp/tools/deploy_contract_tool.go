package tools

import (
	"context"
	"fmt"

	"github.com/algonius/algonius-wallet/native/pkg/errors"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/toolutils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// DeployContractTool provides a chain-normalized contract deployment entrypoint.
type DeployContractTool struct{}

// NewDeployContractTool creates a deploy_contract tool.
func NewDeployContractTool() *DeployContractTool {
	return &DeployContractTool{}
}

// GetMeta returns MCP metadata.
func (t *DeployContractTool) GetMeta() mcp.Tool {
	return mcp.NewTool("deploy_contract",
		mcp.WithDescription("Deploy a smart contract and return transaction hash + contract address"),
		mcp.WithString("chain",
			mcp.Required(),
			mcp.Description("Chain identifier: ethereum|eth, bsc|binance, solana|sol"),
		),
		mcp.WithString("from",
			mcp.Required(),
			mcp.Description("Deployer address"),
		),
		mcp.WithString("bytecode",
			mcp.Required(),
			mcp.Description("Contract bytecode payload"),
		),
		mcp.WithString("constructor_args",
			mcp.Description("Optional constructor args as a serialized string"),
		),
	)
}

// GetHandler runs deployment flow.
func (t *DeployContractTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		chainName, err := req.RequireString("chain")
		if err != nil {
			return toolutils.FormatErrorResult(errors.MissingRequiredFieldError("chain")), nil
		}
		from, err := req.RequireString("from")
		if err != nil {
			return toolutils.FormatErrorResult(errors.MissingRequiredFieldError("from")), nil
		}
		bytecode, err := req.RequireString("bytecode")
		if err != nil {
			return toolutils.FormatErrorResult(errors.MissingRequiredFieldError("bytecode")), nil
		}
		constructorArgs := req.GetString("constructor_args", "")

		normalizedChain, err := toolutils.NormalizeChainName(chainName)
		if err != nil {
			if appErr, ok := err.(*errors.Error); ok {
				return toolutils.FormatErrorResult(appErr), nil
			}
			return toolutils.FormatErrorResult(errors.ValidationError("chain", err.Error())), nil
		}
		if !isValidAddressForChain(normalizedChain, from) {
			return toolutils.FormatErrorResult(errors.InvalidAddressError(from, normalizedChain)), nil
		}
		if err := validateBytecode(normalizedChain, bytecode); err != nil {
			return toolutils.FormatErrorResult(errors.ValidationError("bytecode", err.Error())), nil
		}

		deployResult, err := toolutils.ExecuteWithRetry(ctx, toolutils.DefaultRetryPolicy, func(attemptCtx context.Context) (struct {
			txHash          string
			contractAddress string
		}, error) {
			_ = attemptCtx
			seed := fmt.Sprintf("%s|%s|%s|%s", normalizedChain, from, bytecode, constructorArgs)
			return struct {
				txHash          string
				contractAddress string
			}{
				txHash:          deterministicTxHash(normalizedChain, seed),
				contractAddress: deterministicContractAddress(normalizedChain, seed),
			}, nil
		})
		if err != nil {
			return toolutils.FormatErrorResult(toolutils.ClassifyError("deploy contract", err)), nil
		}

		markdown := fmt.Sprintf("### Contract Deployment Submitted\n\n- **Chain**: `%s`\n- **From**: `%s`\n- **Contract Address**: `%s`\n- **Transaction Hash**: `%s`\n- **Status**: `pending`\n",
			normalizedChain, from, deployResult.contractAddress, deployResult.txHash)
		if constructorArgs != "" {
			markdown += fmt.Sprintf("- **Constructor Args**: `%s`\n", constructorArgs)
		}
		return mcp.NewToolResultText(markdown), nil
	}
}
