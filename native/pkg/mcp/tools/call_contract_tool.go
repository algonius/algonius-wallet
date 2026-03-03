package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/algonius/algonius-wallet/native/pkg/errors"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/toolutils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// CallContractTool performs read-style contract method calls.
type CallContractTool struct{}

// NewCallContractTool creates a call_contract tool.
func NewCallContractTool() *CallContractTool {
	return &CallContractTool{}
}

// GetMeta returns MCP metadata.
func (t *CallContractTool) GetMeta() mcp.Tool {
	return mcp.NewTool("call_contract",
		mcp.WithDescription("Call a contract method and return decoded mock result"),
		mcp.WithString("chain",
			mcp.Required(),
			mcp.Description("Chain identifier: ethereum|eth, bsc|binance, solana|sol"),
		),
		mcp.WithString("contract_address",
			mcp.Required(),
			mcp.Description("Contract/program address"),
		),
		mcp.WithString("method",
			mcp.Required(),
			mcp.Description("Method or instruction name"),
		),
		mcp.WithString("args",
			mcp.Description("Optional method args as a serialized string"),
		),
		mcp.WithString("from",
			mcp.Description("Optional caller address"),
		),
	)
}

// GetHandler handles contract calls with timeout/retry wrappers.
func (t *CallContractTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		chainName, err := req.RequireString("chain")
		if err != nil {
			return toolutils.FormatErrorResult(errors.MissingRequiredFieldError("chain")), nil
		}
		contractAddress, err := req.RequireString("contract_address")
		if err != nil {
			return toolutils.FormatErrorResult(errors.MissingRequiredFieldError("contract_address")), nil
		}
		method, err := req.RequireString("method")
		if err != nil {
			return toolutils.FormatErrorResult(errors.MissingRequiredFieldError("method")), nil
		}
		args := req.GetString("args", "")
		from := req.GetString("from", "")

		normalizedChain, err := toolutils.NormalizeChainName(chainName)
		if err != nil {
			if appErr, ok := err.(*errors.Error); ok {
				return toolutils.FormatErrorResult(appErr), nil
			}
			return toolutils.FormatErrorResult(errors.ValidationError("chain", err.Error())), nil
		}
		if !isValidAddressForChain(normalizedChain, contractAddress) {
			return toolutils.FormatErrorResult(errors.InvalidAddressError(contractAddress, normalizedChain)), nil
		}
		if strings.TrimSpace(method) == "" {
			return toolutils.FormatErrorResult(errors.ValidationError("method", "method cannot be empty")), nil
		}
		if from != "" && !isValidAddressForChain(normalizedChain, from) {
			return toolutils.FormatErrorResult(errors.InvalidAddressError(from, normalizedChain)), nil
		}

		callResult, err := toolutils.ExecuteWithRetry(ctx, toolutils.DefaultRetryPolicy, func(attemptCtx context.Context) (string, error) {
			_ = attemptCtx
			if strings.EqualFold(method, "revert") {
				return "", fmt.Errorf("execution reverted by contract")
			}
			seed := fmt.Sprintf("%s|%s|%s|%s|%s", normalizedChain, contractAddress, method, args, from)
			return deterministicCallResult(normalizedChain, method, seed), nil
		})
		if err != nil {
			return toolutils.FormatErrorResult(toolutils.ClassifyError("call contract", err)), nil
		}

		markdown := fmt.Sprintf("### Contract Call Result\n\n- **Chain**: `%s`\n- **Contract**: `%s`\n- **Method**: `%s`\n- **Result**: `%s`\n",
			normalizedChain, contractAddress, method, callResult)
		if args != "" {
			markdown += fmt.Sprintf("- **Args**: `%s`\n", args)
		}
		if from != "" {
			markdown += fmt.Sprintf("- **Caller**: `%s`\n", from)
		}
		return mcp.NewToolResultText(markdown), nil
	}
}
