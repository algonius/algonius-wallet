package tools

import (
	"context"
	"testing"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

// MockRejectWalletManager is a mock implementation of wallet.IWalletManager for testing reject transaction
type MockRejectWalletManager struct {
	rejectTransactionsFunc func(ctx context.Context, transactionIds []string, reason, details string, notifyUser, auditLog bool) ([]wallet.TransactionRejectionResult, error)
}

func (m *MockRejectWalletManager) CreateWallet(ctx context.Context, chain string) (address string, publicKey string, err error) {
	return "0x123", "0x456", nil
}

func (m *MockRejectWalletManager) ImportWallet(ctx context.Context, mnemonic, password, chainName, derivationPath string) (address string, publicKey string, importedAt int64, err error) {
	return "0x123", "0x456", 1234567890, nil
}

func (m *MockRejectWalletManager) GetBalance(ctx context.Context, address, token string) (string, error) {
	return "0", nil
}

func (m *MockRejectWalletManager) GetStatus(ctx context.Context) (*wallet.WalletStatus, error) {
	return nil, nil
}

func (m *MockRejectWalletManager) SendTransaction(ctx context.Context, chain, from, to, amount, token string) (string, error) {
	return "0xmockhash", nil
}

func (m *MockRejectWalletManager) EstimateGas(ctx context.Context, chain, from, to, amount, token string) (uint64, string, error) {
	return 21000, "20", nil
}

func (m *MockRejectWalletManager) GetPendingTransactions(ctx context.Context, chain, address, transactionType string, limit, offset int) ([]*wallet.PendingTransaction, error) {
	return []*wallet.PendingTransaction{}, nil
}

func (m *MockRejectWalletManager) RejectTransactions(ctx context.Context, transactionIds []string, reason, details string, notifyUser, auditLog bool) ([]wallet.TransactionRejectionResult, error) {
	if m.rejectTransactionsFunc != nil {
		return m.rejectTransactionsFunc(ctx, transactionIds, reason, details, notifyUser, auditLog)
	}
	return []wallet.TransactionRejectionResult{}, nil
}

func (m *MockRejectWalletManager) GetTransactionHistory(ctx context.Context, address string, fromBlock, toBlock *uint64, limit, offset int) ([]*wallet.HistoricalTransaction, error) {
	return []*wallet.HistoricalTransaction{}, nil
}

func TestRejectTransactionTool_GetMeta(t *testing.T) {
	mockManager := &MockRejectWalletManager{}
	tool := NewRejectTransactionTool(mockManager)

	meta := tool.GetMeta()

	if meta.GetName() != "reject_transaction" {
		t.Errorf("Expected tool name 'reject_transaction', got '%s'", meta.GetName())
	}

	if meta.Description != "Reject pending transactions by ID with specified reasons and optional notifications" {
		t.Errorf("Expected description 'Reject pending transactions by ID with specified reasons and optional notifications', got '%s'", meta.Description)
	}
}

func TestRejectTransactionTool_isValidReason(t *testing.T) {
	mockManager := &MockRejectWalletManager{}
	tool := NewRejectTransactionTool(mockManager)

	validReasons := []string{
		"suspicious_activity",
		"high_gas_fee",
		"user_request",
		"security_concern",
		"duplicate_transaction",
	}

	for _, reason := range validReasons {
		if !tool.isValidReason(reason) {
			t.Errorf("Expected reason '%s' to be valid", reason)
		}
	}

	invalidReasons := []string{
		"invalid_reason",
		"",
		"unknown",
	}

	for _, reason := range invalidReasons {
		if tool.isValidReason(reason) {
			t.Errorf("Expected reason '%s' to be invalid", reason)
		}
	}
}

func TestRejectTransactionTool_Handler_Success(t *testing.T) {
	// Mock the RejectTransactions function to return successful results
	mockManager := &MockRejectWalletManager{
		rejectTransactionsFunc: func(ctx context.Context, transactionIds []string, reason, details string, notifyUser, auditLog bool) ([]wallet.TransactionRejectionResult, error) {
			results := make([]wallet.TransactionRejectionResult, len(transactionIds))
			for i, id := range transactionIds {
				results[i] = wallet.TransactionRejectionResult{
					TransactionHash: id,
					Success:         true,
				}
			}
			return results, nil
		},
	}

	tool := NewRejectTransactionTool(mockManager)
	handler := tool.GetHandler()

	// Create a proper CallToolRequest
	params := mcp.CallToolParams{
		Name: "reject_transaction",
		Arguments: map[string]interface{}{
			"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
			"reason":          "suspicious_activity",
		},
	}
	
	req := mcp.CallToolRequest{}
	req.Params = params

	result, err := handler(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.NotNil(t, result.Content)
}

func TestRejectTransactionTool_Handler_MissingParameters(t *testing.T) {
	mockManager := &MockRejectWalletManager{}
	tool := NewRejectTransactionTool(mockManager)
	handler := tool.GetHandler()

	// Test with missing transaction_ids
	params := mcp.CallToolParams{
		Name: "reject_transaction",
		Arguments: map[string]interface{}{
			"reason": "suspicious_activity",
		},
	}
	
	req := mcp.CallToolRequest{}
	req.Params = params

	result, err := handler(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsError)

	// Test with missing reason
	params = mcp.CallToolParams{
		Name: "reject_transaction",
		Arguments: map[string]interface{}{
			"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		},
	}
	
	req = mcp.CallToolRequest{}
	req.Params = params

	result, err = handler(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestRejectTransactionTool_Handler_InvalidReason(t *testing.T) {
	mockManager := &MockRejectWalletManager{}
	tool := NewRejectTransactionTool(mockManager)
	handler := tool.GetHandler()

	// Test with invalid reason
	params := mcp.CallToolParams{
		Name: "reject_transaction",
		Arguments: map[string]interface{}{
			"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
			"reason":          "invalid_reason",
		},
	}
	
	req := mcp.CallToolRequest{}
	req.Params = params

	result, err := handler(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestRejectTransactionTool_Handler_EmptyTransactionIds(t *testing.T) {
	mockManager := &MockRejectWalletManager{}
	tool := NewRejectTransactionTool(mockManager)
	handler := tool.GetHandler()

	// Test with empty transaction_ids
	params := mcp.CallToolParams{
		Name: "reject_transaction",
		Arguments: map[string]interface{}{
			"transaction_ids": "",
			"reason":          "suspicious_activity",
		},
	}
	
	req := mcp.CallToolRequest{}
	req.Params = params

	result, err := handler(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestRejectTransactionTool_Handler_MultipleTransactionIds(t *testing.T) {
	// Mock the RejectTransactions function to return successful results
	mockManager := &MockRejectWalletManager{
		rejectTransactionsFunc: func(ctx context.Context, transactionIds []string, reason, details string, notifyUser, auditLog bool) ([]wallet.TransactionRejectionResult, error) {
			results := make([]wallet.TransactionRejectionResult, len(transactionIds))
			for i, id := range transactionIds {
				results[i] = wallet.TransactionRejectionResult{
					TransactionHash: id,
					Success:         true,
				}
			}
			return results, nil
		},
	}

	tool := NewRejectTransactionTool(mockManager)
	handler := tool.GetHandler()

	// Test with multiple transaction IDs
	params := mcp.CallToolParams{
		Name: "reject_transaction",
		Arguments: map[string]interface{}{
			"transaction_ids": "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456,0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"reason":          "high_gas_fee",
			"details":         "Gas fee exceeds 0.01 ETH threshold",
			"notify_user":     true,
			"audit_log":       true,
		},
	}
	
	req := mcp.CallToolRequest{}
	req.Params = params

	result, err := handler(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.NotNil(t, result.Content)
}