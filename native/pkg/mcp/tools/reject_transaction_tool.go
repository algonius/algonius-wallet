// Package tools provides MCP tool implementations for the Algonius Native Host.
package tools

import (
	"context"
	
	"fmt"
	"strings"

	"github.com/algonius/algonius-wallet/native/pkg/errors"
	"github.com/algonius/algonius-wallet/native/pkg/mcp/toolutils"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RejectTransactionTool implements the MCP "reject_transaction" tool for rejecting pending transactions.
type RejectTransactionTool struct {
	manager wallet.IWalletManager
}

// NewRejectTransactionTool constructs a RejectTransactionTool with the given wallet manager.
func NewRejectTransactionTool(manager wallet.IWalletManager) *RejectTransactionTool {
	return &RejectTransactionTool{manager: manager}
}

// GetMeta returns the MCP tool definition for "reject_transaction" as per the documented API schema.
func (t *RejectTransactionTool) GetMeta() mcp.Tool {
	return mcp.NewTool("reject_transaction",
		mcp.WithDescription("Reject pending transactions by ID with specified reasons and optional notifications"),
		mcp.WithString("transaction_ids",
			mcp.Required(),
			mcp.Description("Comma-separated list of transaction hashes to reject"),
		),
		mcp.WithString("reason",
			mcp.Required(),
			mcp.Description("Reason for rejection (e.g., 'suspicious_activity', 'high_gas_fee', 'user_request', 'security_concern', 'duplicate_transaction')"),
		),
		mcp.WithString("details",
			mcp.Description("Additional details about the rejection reason"),
		),
		mcp.WithBoolean("notify_user",
			mcp.Description("Whether to send notification to user about the rejection (default: false)"),
		),
		mcp.WithBoolean("audit_log",
			mcp.Description("Whether to log the rejection in audit log (default: true)"),
		),
	)
}

// GetHandler returns the handler function for the "reject_transaction" tool.
// The handler rejects the specified transactions with provided reasons and optional notifications.
func (t *RejectTransactionTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract and validate transaction_ids parameter
		transactionIdsStr, err := req.RequireString("transaction_ids")
		if err != nil {
			toolErr := errors.MissingRequiredFieldError("transaction_ids")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Extract and validate reason parameter
		reason, err := req.RequireString("reason")
		if err != nil {
			toolErr := errors.MissingRequiredFieldError("reason")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Parse transaction IDs from comma-separated string
		transactionIds := strings.Split(transactionIdsStr, ",")
		var cleanedIds []string
		for _, id := range transactionIds {
			cleanedId := strings.TrimSpace(id)
			if cleanedId != "" {
				cleanedIds = append(cleanedIds, cleanedId)
			}
		}

		if len(cleanedIds) == 0 {
			toolErr := errors.ValidationError("transaction_ids", "no valid transaction IDs provided")
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Validate reason
		if !t.isValidReason(reason) {
			toolErr := errors.ValidationError("reason", fmt.Sprintf("invalid reason: %s. Valid reasons: suspicious_activity, high_gas_fee, user_request, security_concern, duplicate_transaction", reason))
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Extract optional parameters
		details := req.GetString("details", "")
		notifyUser := req.GetBool("notify_user", false)
		auditLog := req.GetBool("audit_log", true)

		// Reject transactions
		rejectionResults, err := t.manager.RejectTransactions(ctx, cleanedIds, reason, details, notifyUser, auditLog)
		if err != nil {
			toolErr := errors.InternalError("reject transactions", err)
			return toolutils.FormatErrorResult(toolErr), nil
		}

		// Format response as markdown
		markdown := t.formatRejectionResponse(rejectionResults, reason, details, notifyUser, auditLog)

		return mcp.NewToolResultText(markdown), nil
	}
}


// isValidReason validates the rejection reason
func (t *RejectTransactionTool) isValidReason(reason string) bool {
	validReasons := map[string]bool{
		"suspicious_activity":    true,
		"high_gas_fee":          true,
		"user_request":          true,
		"security_concern":      true,
		"duplicate_transaction": true,
	}
	return validReasons[strings.ToLower(reason)]
}

// formatRejectionResponse formats the rejection results as markdown
func (t *RejectTransactionTool) formatRejectionResponse(results []wallet.TransactionRejectionResult, reason, details string, notifyUser, auditLog bool) string {
	markdown := "### Transaction Rejection Results\n\n"

	successCount := 0
	failureCount := 0
	
	// Count successes and failures
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	// Summary
	markdown += fmt.Sprintf("**Summary**: %d transaction(s) processed\n", len(results))
	markdown += fmt.Sprintf("- ✅ Successfully rejected: %d\n", successCount)
	markdown += fmt.Sprintf("- ❌ Failed to reject: %d\n\n", failureCount)

	// Rejection details
	markdown += "**Rejection Details**:\n"
	markdown += fmt.Sprintf("- **Reason**: `%s`\n", reason)
	if details != "" {
		markdown += fmt.Sprintf("- **Details**: %s\n", details)
	}
	markdown += fmt.Sprintf("- **User Notification**: %t\n", notifyUser)
	markdown += fmt.Sprintf("- **Audit Logging**: %t\n\n", auditLog)

	// Individual results
	markdown += "#### Individual Results\n\n"
	for i, result := range results {
		status := "✅ SUCCESS"
		if !result.Success {
			status = "❌ FAILED"
		}

		markdown += fmt.Sprintf("**Transaction %d** %s\n", i+1, status)
		markdown += fmt.Sprintf("- **Hash**: `%s`\n", result.TransactionHash)
		
		if result.Success {
			markdown += fmt.Sprintf("- **Rejected At**: `%s`\n", result.RejectedAt.Format("2006-01-02 15:04:05"))
			if result.AuditLogId != "" {
				markdown += fmt.Sprintf("- **Audit Log ID**: `%s`\n", result.AuditLogId)
			}
		} else {
			markdown += fmt.Sprintf("- **Error**: %s\n", result.ErrorMessage)
		}
		markdown += "\n"
	}

	// Additional information
	if notifyUser && successCount > 0 {
		markdown += "---\n"
		markdown += "**User Notification**: Notifications have been sent for successfully rejected transactions.\n"
	}

	if auditLog && successCount > 0 {
		markdown += "---\n"
		markdown += "**Audit Trail**: All successful rejections have been logged for security auditing.\n"
	}

	// Security notice
	if successCount > 0 {
		markdown += "---\n"
		markdown += "⚠️ **Security Notice**: Rejected transactions cannot be recovered. This action has been permanently logged.\n"
	}

	return markdown
}
