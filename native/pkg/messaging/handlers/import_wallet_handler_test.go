// SPDX-License-Identifier: Apache-2.0
package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/algonius/algonius-wallet/native/pkg/messaging"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
)

// MockWalletManagerForImport is a mock implementation for testing import_wallet
type MockWalletManagerForImport struct {
	ShouldFail    bool
	FailureReason string
}

func (m *MockWalletManagerForImport) CreateWallet(ctx context.Context, chain string) (address string, publicKey string, err error) {
	return "0x123", "0x456", nil
}

func (m *MockWalletManagerForImport) ImportWallet(ctx context.Context, mnemonic, password, chainName, derivationPath string) (address string, publicKey string, importedAt int64, err error) {
	if m.ShouldFail {
		return "", "", 0, &ImportError{Reason: m.FailureReason}
	}
	return "0x1234567890abcdef1234567890abcdef12345678", "0x04abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", 1234567890, nil
}

func (m *MockWalletManagerForImport) GetBalance(ctx context.Context, address, token string) (string, error) {
	return "0", nil
}

func (m *MockWalletManagerForImport) GetStatus(ctx context.Context) (*wallet.WalletStatus, error) {
	return nil, nil
}

func (m *MockWalletManagerForImport) SendTransaction(ctx context.Context, chain, from, to, amount, token string) (string, error) {
	return "0xmockhash", nil
}

func (m *MockWalletManagerForImport) EstimateGas(ctx context.Context, chain, from, to, amount, token string) (uint64, string, error) {
	return 21000, "20", nil
}

func (m *MockWalletManagerForImport) GetPendingTransactions(ctx context.Context, chain, address, transactionType string, limit, offset int) ([]*wallet.PendingTransaction, error) {
	return []*wallet.PendingTransaction{}, nil
}

func (m *MockWalletManagerForImport) RejectTransactions(ctx context.Context, transactionIds []string, reason, details string, notifyUser, auditLog bool) ([]wallet.TransactionRejectionResult, error) {
	return []wallet.TransactionRejectionResult{}, nil
}

func (m *MockWalletManagerForImport) GetTransactionHistory(ctx context.Context, address string, fromBlock, toBlock *uint64, limit, offset int) ([]*wallet.HistoricalTransaction, error) {
	return []*wallet.HistoricalTransaction{}, nil
}

// ImportError helps simulate specific error types
type ImportError struct {
	Reason string
}

func (e *ImportError) Error() string {
	return e.Reason
}

func TestCreateImportWalletHandler_Success(t *testing.T) {
	mockManager := &MockWalletManagerForImport{}
	handler := CreateImportWalletHandler(mockManager)

	params := ImportWalletParams{
		Mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		Password: "password123",
		Chain:    "ethereum",
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal params: %v", err)
	}

	request := messaging.RpcRequest{
		ID:     "test-id",
		Method: "import_wallet",
		Params: paramsJSON,
	}

	response, err := handler(request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if response.Error != nil {
		t.Fatalf("Expected success but got error: %v", response.Error)
	}

	if response.Result == nil {
		t.Fatal("Expected result but got nil")
	}

	var result ImportWalletResult
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result.Address == "" {
		t.Error("Expected address but got empty string")
	}
	if result.PublicKey == "" {
		t.Error("Expected public key but got empty string")
	}
	if result.ImportedAt == 0 {
		t.Error("Expected imported timestamp but got 0")
	}
}

func TestCreateImportWalletHandler_MissingParameters(t *testing.T) {
	mockManager := &MockWalletManagerForImport{}
	handler := CreateImportWalletHandler(mockManager)

	tests := []struct {
		name           string
		params         ImportWalletParams
		expectedError  int
	}{
		{
			name: "missing mnemonic",
			params: ImportWalletParams{
				Password: "password123",
				Chain:    "ethereum",
			},
			expectedError: ErrInvalidMnemonic,
		},
		{
			name: "missing password",
			params: ImportWalletParams{
				Mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
				Chain:    "ethereum",
			},
			expectedError: ErrWeakPassword,
		},
		{
			name: "missing chain",
			params: ImportWalletParams{
				Mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
				Password: "password123",
			},
			expectedError: ErrUnsupportedChain,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paramsJSON, err := json.Marshal(tt.params)
			if err != nil {
				t.Fatalf("Failed to marshal params: %v", err)
			}

			request := messaging.RpcRequest{
				ID:     "test-id",
				Method: "import_wallet",
				Params: paramsJSON,
			}

			response, err := handler(request)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if response.Error == nil {
				t.Fatal("Expected error but got success")
			}

			if response.Error.Code != tt.expectedError {
				t.Errorf("Expected error code %d but got %d", tt.expectedError, response.Error.Code)
			}
		})
	}
}

func TestCreateImportWalletHandler_InvalidJson(t *testing.T) {
	mockManager := &MockWalletManagerForImport{}
	handler := CreateImportWalletHandler(mockManager)

	request := messaging.RpcRequest{
		ID:     "test-id",
		Method: "import_wallet",
		Params: json.RawMessage(`{"invalid":"json"`), // Missing closing brace
	}

	response, err := handler(request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if response.Error == nil {
		t.Fatal("Expected error but got success")
	}

	if response.Error.Code != -32602 {
		t.Errorf("Expected error code -32602 but got %d", response.Error.Code)
	}
}

func TestCreateImportWalletHandler_WalletManagerErrors(t *testing.T) {
	tests := []struct {
		name            string
		failureReason   string
		expectedError   int
	}{
		{
			name:          "invalid mnemonic error",
			failureReason: "invalid mnemonic phrase",
			expectedError: ErrInvalidMnemonic,
		},
		{
			name:          "weak password error",
			failureReason: "weak password provided",
			expectedError: ErrWeakPassword,
		},
		{
			name:          "unsupported chain error",
			failureReason: "unsupported chain bitcoin",
			expectedError: ErrUnsupportedChain,
		},
		{
			name:          "wallet already exists error",
			failureReason: "wallet already exists",
			expectedError: ErrWalletAlreadyExists,
		},
		{
			name:          "storage encryption failed error",
			failureReason: "storage encryption failed",
			expectedError: ErrStorageEncryptionFailed,
		},
		{
			name:          "generic error",
			failureReason: "some other error",
			expectedError: -32000, // Default server error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := &MockWalletManagerForImport{
				ShouldFail:    true,
				FailureReason: tt.failureReason,
			}
			handler := CreateImportWalletHandler(mockManager)

			params := ImportWalletParams{
				Mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
				Password: "password123",
				Chain:    "ethereum",
			}

			paramsJSON, err := json.Marshal(params)
			if err != nil {
				t.Fatalf("Failed to marshal params: %v", err)
			}

			request := messaging.RpcRequest{
				ID:     "test-id",
				Method: "import_wallet",
				Params: paramsJSON,
			}

			response, err := handler(request)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if response.Error == nil {
				t.Fatal("Expected error but got success")
			}

			if response.Error.Code != tt.expectedError {
				t.Errorf("Expected error code %d but got %d", tt.expectedError, response.Error.Code)
			}
		})
	}
}

func TestCreateImportWalletHandler_NilParams(t *testing.T) {
	mockManager := &MockWalletManagerForImport{}
	handler := CreateImportWalletHandler(mockManager)

	request := messaging.RpcRequest{
		ID:     "test-id",
		Method: "import_wallet",
		Params: nil,
	}

	response, err := handler(request)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if response.Error == nil {
		t.Fatal("Expected error but got success")
	}

	if response.Error.Code != ErrInvalidMnemonic {
		t.Errorf("Expected error code %d but got %d", ErrInvalidMnemonic, response.Error.Code)
	}
}