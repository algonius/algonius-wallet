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
	mockHistoricalTransactions []*wallet.HistoricalTransaction
	shouldReturnError          bool
}

func (m *MockWalletManagerWithHistory) CreateWallet(ctx context.Context, chain string, password string) (address string, publicKey string, mnemonic string, err error) {
	return "0x123", "0x456", "", nil
}

func (m *MockWalletManagerWithHistory) ImportWallet(ctx context.Context, mnemonic, password, chainName, derivationPath string) (address string, publicKey string, importedAt int64, err error) {
	return "0x123", "0x456", 1234567890, nil
}

func (m *MockWalletManagerWithHistory) GetBalance(ctx context.Context, address, token string) (string, error) {
	return "0", nil
}

func (m *MockWalletManagerWithHistory) GetStatus(ctx context.Context) (*wallet.WalletStatus, error) {
	return nil, nil
}

func (m *MockWalletManagerWithHistory) SendTransaction(ctx context.Context, chain, from, to, amount, token string) (string, error) {
	return "0xmockhash", nil
}

func (m *MockWalletManagerWithHistory) EstimateGas(ctx context.Context, chain, from, to, amount, token string) (uint64, string, error) {
	return 21000, "20", nil
}

func (m *MockWalletManagerWithHistory) GetPendingTransactions(ctx context.Context, chain, address, transactionType string, limit, offset int) ([]*wallet.PendingTransaction, error) {
	return []*wallet.PendingTransaction{}, nil
}

func (m *MockWalletManagerWithHistory) RejectTransactions(ctx context.Context, transactionIds []string, reason, details string, notifyUser, auditLog bool) ([]wallet.TransactionRejectionResult, error) {
	return []wallet.TransactionRejectionResult{}, nil
}

func (m *MockWalletManagerWithHistory) GetTransactionHistory(ctx context.Context, address string, fromBlock, toBlock *uint64, limit, offset int) ([]*wallet.HistoricalTransaction, error) {
	if m.shouldReturnError {
		return nil, assert.AnError
	}
	return m.mockHistoricalTransactions, nil
}

func (m *MockWalletManagerWithHistory) GetAccounts(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

func (m *MockWalletManagerWithHistory) AddPendingTransaction(ctx context.Context, tx *wallet.PendingTransaction) error {
	return nil
}

func (m *MockWalletManagerWithHistory) UnlockWallet(password string) error {
	return nil
}

func (m *MockWalletManagerWithHistory) LockWallet() {
}

func (m *MockWalletManagerWithHistory) IsUnlocked() bool {
	return true
}

func (m *MockWalletManagerWithHistory) HasWallet() bool {
	return true
}

func (m *MockWalletManagerWithHistory) GetCurrentWallet() *wallet.WalletStatus {
	return &wallet.WalletStatus{}
}

func TestGetTransactionHistoryToolMeta(t *testing.T) {
	mockManager := &MockWalletManagerWithHistory{}
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
	mockManager := &MockWalletManagerWithHistory{}
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
	mockManager := &MockWalletManagerWithHistory{}
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
