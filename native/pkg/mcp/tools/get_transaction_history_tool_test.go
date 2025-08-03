package tools

import (
	"context"
	"testing"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockWalletManagerWithHistory is a mock implementation of wallet.IWalletManager for testing transaction history
type MockWalletManagerWithHistory struct {
	*wallet.MockWalletManager
	mockHistoricalTransactions []*wallet.HistoricalTransaction
	shouldReturnError          bool
}

func (m *MockWalletManagerWithHistory) GetTransactionHistory(ctx context.Context, address string, fromBlock, toBlock *uint64, limit, offset int) ([]*wallet.HistoricalTransaction, error) {
	if m.shouldReturnError {
		return nil, assert.AnError
	}
	return m.mockHistoricalTransactions, nil
}

func TestGetTransactionHistoryToolMeta(t *testing.T) {
	mockManager := &wallet.MockWalletManager{}
	tool := NewGetTransactionHistoryTool(mockManager)

	meta := tool.GetMeta()

	assert.Equal(t, "get_transaction_history", meta.Name)
	assert.Equal(t, "Query transaction history for a wallet address", meta.Description)

	// Check that all expected parameters are present
	assert.Contains(t, meta.InputSchema.Properties, "address")
	assert.Contains(t, meta.InputSchema.Properties, "limit")
	assert.Contains(t, meta.InputSchema.Properties, "from_block")
	assert.Contains(t, meta.InputSchema.Properties, "to_block")

	// Check that address is required
	required := meta.InputSchema.Required
	assert.Contains(t, required, "address")
}

func TestGetTransactionHistoryToolCreation(t *testing.T) {
	mockManager := &wallet.MockWalletManager{}
	tool := NewGetTransactionHistoryTool(mockManager)

	require.NotNil(t, tool)
	require.NotNil(t, tool.manager)

	// Test that the tool has the expected methods
	meta := tool.GetMeta()
	require.NotNil(t, meta)

	handler := tool.GetHandler()
	require.NotNil(t, handler)
}

func TestGetTransactionHistoryToolHandler_Success(t *testing.T) {
	// Create mock historical transactions
	baseTime := time.Now()
	mockTxs := []*wallet.HistoricalTransaction{
		{
			Hash:           "0x123abc456def789ghi012jkl345mno678pqr901stu234vwx567yza890bcd123efg",
			Chain:          "ethereum",
			BlockNumber:    18500100,
			From:           "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			To:             "0x8ba1f109551bD432803012645Hac136c22C4F9B",
			Value:          "1.5",
			TokenSymbol:    "ETH",
			Type:           "transfer",
			Status:         "confirmed",
			TransactionFee: "0.00042",
			Timestamp:      baseTime.Add(-2 * time.Hour),
			Confirmations:  50,
		},
		{
			Hash:            "0x789def012abc345ghi678jkl901mno234pqr567stu890vwx123yza456bcd789efg",
			Chain:           "ethereum",
			BlockNumber:     18500095,
			From:            "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			To:              "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984",
			Value:           "0",
			TokenSymbol:     "UNI",
			Type:            "contract_call",
			Status:          "confirmed",
			TransactionFee:  "0.001177776",
			Timestamp:       baseTime.Add(-4 * time.Hour),
			Confirmations:   55,
			ContractAddress: "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984",
			MethodName:      "transfer",
		},
	}

	mockManager := &MockWalletManagerWithHistory{
		MockWalletManager:          &wallet.MockWalletManager{},
		mockHistoricalTransactions: mockTxs,
	}
	tool := NewGetTransactionHistoryTool(mockManager)
	handler := tool.GetHandler()

	// Create proper request structure
	params := mcp.CallToolParams{
		Name: "get_transaction_history",
		Arguments: map[string]interface{}{
			"address": "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			"limit":   10,
		},
	}

	req := mcp.CallToolRequest{}
	req.Params = params

	// Execute the handler
	result, err := handler(context.Background(), req)

	// Verify the result
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
	require.NotNil(t, result.Content)
	require.Len(t, result.Content, 1)

	// Check that the markdown output contains expected content
	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "result should contain text content")
	markdown := textContent.Text
	assert.Contains(t, markdown, "### Transaction History")
	assert.Contains(t, markdown, "Found 2 transactions")
	assert.Contains(t, markdown, "0x123abc456def789ghi012jkl345mno678pqr901stu234vwx567yza890bcd123efg")
	assert.Contains(t, markdown, "0x789def012abc345ghi678jkl901mno234pqr567stu890vwx123yza456bcd789efg")
	assert.Contains(t, markdown, "ethereum")
	assert.Contains(t, markdown, "transfer")
	assert.Contains(t, markdown, "contract_call")
}

func TestGetTransactionHistoryToolHandler_NoTransactions(t *testing.T) {
	mockManager := &MockWalletManagerWithHistory{
		MockWalletManager:          &wallet.MockWalletManager{},
		mockHistoricalTransactions: []*wallet.HistoricalTransaction{},
	}
	tool := NewGetTransactionHistoryTool(mockManager)
	handler := tool.GetHandler()

	// Create proper request structure
	params := mcp.CallToolParams{
		Name: "get_transaction_history",
		Arguments: map[string]interface{}{
			"address": "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
		},
	}

	req := mcp.CallToolRequest{}
	req.Params = params

	// Execute the handler
	result, err := handler(context.Background(), req)

	// Verify the result
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
	require.NotNil(t, result.Content)
	require.Len(t, result.Content, 1)

	// Check that the markdown output indicates no transactions
	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "result should contain text content")
	markdown := textContent.Text
	assert.Contains(t, markdown, "### Transaction History")
	assert.Contains(t, markdown, "No transactions found")
}

func TestGetTransactionHistoryToolHandler_MissingAddress(t *testing.T) {
	mockManager := &wallet.MockWalletManager{}
	tool := NewGetTransactionHistoryTool(mockManager)
	handler := tool.GetHandler()

	// Create request without required address parameter
	params := mcp.CallToolParams{
		Name:      "get_transaction_history",
		Arguments: map[string]interface{}{},
	}

	req := mcp.CallToolRequest{}
	req.Params = params

	// Execute the handler
	result, err := handler(context.Background(), req)

	// Verify that it returns an error
	require.NoError(t, err) // Handler shouldn't return Go error
	require.NotNil(t, result)
	require.True(t, result.IsError)
}

func TestGetTransactionHistoryToolHandler_WithBlockRange(t *testing.T) {
	mockTxs := []*wallet.HistoricalTransaction{
		{
			Hash:           "0x123abc456def789ghi012jkl345mno678pqr901stu234vwx567yza890bcd123efg",
			Chain:          "ethereum",
			BlockNumber:    18500100,
			From:           "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			To:             "0x8ba1f109551bD432803012645Hac136c22C4F9B",
			Value:          "1.5",
			TokenSymbol:    "ETH",
			Type:           "transfer",
			Status:         "confirmed",
			TransactionFee: "0.00042",
			Timestamp:      time.Now().Add(-2 * time.Hour),
			Confirmations:  50,
		},
	}

	mockManager := &MockWalletManagerWithHistory{
		MockWalletManager:          &wallet.MockWalletManager{},
		mockHistoricalTransactions: mockTxs,
	}
	tool := NewGetTransactionHistoryTool(mockManager)
	handler := tool.GetHandler()

	// Create request with block range parameters
	params := mcp.CallToolParams{
		Name: "get_transaction_history",
		Arguments: map[string]interface{}{
			"address":    "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
			"from_block": 18500000,
			"to_block":   18500200,
			"limit":      5,
		},
	}

	req := mcp.CallToolRequest{}
	req.Params = params

	// Execute the handler
	result, err := handler(context.Background(), req)

	// Verify the result
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
	require.NotNil(t, result.Content)
	require.Len(t, result.Content, 1)

	// Check that the markdown output contains block range info
	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "result should contain text content")
	markdown := textContent.Text
	assert.Contains(t, markdown, "### Transaction History")
	assert.Contains(t, markdown, "Block Range")
	assert.Contains(t, markdown, "from 18500000")
	assert.Contains(t, markdown, "to 18500200")
}

func TestGetTransactionHistoryToolHandler_ManagerError(t *testing.T) {
	mockManager := &MockWalletManagerWithHistory{
		MockWalletManager:  &wallet.MockWalletManager{},
		shouldReturnError: true,
	}
	tool := NewGetTransactionHistoryTool(mockManager)
	handler := tool.GetHandler()

	// Create request with required parameters
	params := mcp.CallToolParams{
		Name: "get_transaction_history",
		Arguments: map[string]interface{}{
			"address": "0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8",
		},
	}

	req := mcp.CallToolRequest{}
	req.Params = params

	// Execute the handler
	result, err := handler(context.Background(), req)

	// Verify that it returns an error response
	require.NoError(t, err) // Handler shouldn't return Go error
	require.NotNil(t, result)
	require.True(t, result.IsError)
}
