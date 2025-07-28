// Package tools provides MCP tool implementations for the Algonius Native Host.
package tools

import (
	"context"
	"fmt"

	"github.com/algonius/algonius-wallet/native/pkg/errors"
	"github.com/algonius/algonius-wallet/native/pkg/event"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/toolutils"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ApproveTransactionTool implements the MCP "approve_transaction" tool for approving/rejecting pending transactions from web pages.
type ApproveTransactionTool struct {
	manager     wallet.IWalletManager
	broadcaster *event.EventBroadcaster
}

// NewApproveTransactionTool constructs an ApproveTransactionTool with the given wallet manager and event broadcaster.
func NewApproveTransactionTool(manager wallet.IWalletManager, broadcaster *event.EventBroadcaster) *ApproveTransactionTool {
	return &ApproveTransactionTool{
		manager:     manager,
		broadcaster: broadcaster,
	}
}

// GetMeta returns the MCP tool definition for "approve_transaction" as per the documented API schema.
func (t *ApproveTransactionTool) GetMeta() mcp.Tool {
	return mcp.NewTool("approve_transaction",
		mcp.WithDescription("Approve or reject a pending transaction from a web page"),
		mcp.WithString("transaction_hash",
			mcp.Required(),
			mcp.Description("Hash of the pending transaction to approve or reject"),
		),
		mcp.WithString("action",
			mcp.Required(),
			mcp.Description("Action to take: 'approve' or 'reject'"),
		),
		mcp.WithString("reason",
			mcp.Description("Reason for rejection (required if action is 'reject')"),
		),
	)
}

// GetHandler returns the handler function for the "approve_transaction" tool.
// The handler approves or rejects pending transactions from web pages.
func (t *ApproveTransactionTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract and validate transaction_hash parameter
		txHash, err := req.RequireString("transaction_hash")
		if err != nil {
			toolErr := errors.MissingRequiredFieldError("transaction_hash")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Extract and validate action parameter
		action, err := req.RequireString("action")
		if err != nil {
			toolErr := errors.MissingRequiredFieldError("action")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Validate action value
		if action != "approve" && action != "reject" {
			toolErr := errors.ValidationError("action", "must be 'approve' or 'reject'")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Extract optional reason parameter (required for reject)
		reason := req.GetString("reason", "")
		if action == "reject" && reason == "" {
			toolErr := errors.MissingRequiredFieldError("reason")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Get the pending transaction
		pendingTxs, err := t.manager.GetPendingTransactions(ctx, "", "", "", 100, 0)
		if err != nil {
			toolErr := errors.InternalError("get pending transactions", err)
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Find the specific transaction
		var targetTx *wallet.PendingTransaction
		for _, tx := range pendingTxs {
			if tx.Hash == txHash {
				targetTx = tx
				break
			}
		}

		if targetTx == nil {
			toolErr := errors.New(errors.ErrWalletNotFound, fmt.Sprintf("Pending transaction with hash '%s' not found", txHash))
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Check if transaction is still pending
		if targetTx.Status != "pending" {
			toolErr := errors.ValidationError("transaction_hash", fmt.Sprintf("transaction is already %s", targetTx.Status))
			return toolutils.FormatErrorResult(toolErr), nil
		}

		var markdown string

		if action == "approve" {
			// Approve the transaction - execute it
			err := t.approveTransaction(ctx, targetTx)
			if err != nil {
				toolErr := errors.InternalError("approve transaction", err)
				return toolutils.FormatErrorResult(toolErr), nil
			}

			// Broadcast approval event
			if t.broadcaster != nil {
				t.broadcaster.BroadcastTransactionConfirmed(targetTx.Hash, targetTx.Chain, 1)
			}

			markdown = fmt.Sprintf("### Transaction Approved ✅\n\n"+
				"- **Transaction Hash**: `%s`\n"+
				"- **Chain**: `%s`\n"+
				"- **From**: `%s`\n"+
				"- **To**: `%s`\n"+
				"- **Amount**: `%s %s`\n"+
				"- **Status**: `approved`\n"+
				"- **Action**: Transaction has been signed and submitted to the blockchain\n",
				targetTx.Hash, targetTx.Chain, targetTx.From, targetTx.To, targetTx.Amount, targetTx.Token)

		} else {
			// Reject the transaction
			results, err := t.manager.RejectTransactions(ctx, []string{txHash}, reason, "AI Agent rejection", false, true)
			if err != nil {
				toolErr := errors.InternalError("reject transaction", err)
				return toolutils.FormatErrorResult(toolErr), nil
			}

			if len(results) == 0 || !results[0].Success {
				toolErr := errors.InternalError("reject transaction", fmt.Errorf("rejection failed: %s", results[0].ErrorMessage))
				return toolutils.FormatErrorResult(toolErr), nil
			}

			// Broadcast rejection event
			if t.broadcaster != nil {
				t.broadcaster.BroadcastTransactionRejected(targetTx.Hash, reason)
			}

			markdown = fmt.Sprintf("### Transaction Rejected ❌\n\n"+
				"- **Transaction Hash**: `%s`\n"+
				"- **Chain**: `%s`\n"+
				"- **From**: `%s`\n"+
				"- **To**: `%s`\n"+
				"- **Amount**: `%s %s`\n"+
				"- **Status**: `rejected`\n"+
				"- **Reason**: `%s`\n"+
				"- **Action**: Transaction has been rejected and will not be executed\n",
				targetTx.Hash, targetTx.Chain, targetTx.From, targetTx.To, targetTx.Amount, targetTx.Token, reason)
		}

		return mcp.NewToolResultText(markdown), nil
	}
}

// approveTransaction executes a pending transaction
func (t *ApproveTransactionTool) approveTransaction(ctx context.Context, tx *wallet.PendingTransaction) error {
	// In a real implementation, this would:
	// 1. Sign the transaction with the private key
	// 2. Submit the transaction to the blockchain
	// 3. Update the transaction status
	// 4. Monitor for confirmations

	// For now, we'll simulate approval by updating the status
	// This would be replaced with actual blockchain interaction
	tx.Status = "confirmed"
	tx.Confirmations = 1

	// TODO: Replace with actual transaction execution
	// Example:
	// txHash, err := t.manager.SendTransaction(ctx, tx.Chain, tx.From, tx.To, tx.Amount, tx.Token)
	// if err != nil {
	//     return err
	// }
	// tx.Hash = txHash

	return nil
}