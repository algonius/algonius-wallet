// SPDX-License-Identifier: Apache-2.0
package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/algonius/algonius-wallet/native/pkg/messaging"
	"github.com/algonius/algonius-wallet/native/pkg/wallet"
	"github.com/stretchr/testify/mock"
)

// MockWalletManagerForImport is a mock implementation for testing import_wallet
type MockWalletManagerForImport struct {
	*wallet.MockWalletManager
	ShouldFail    bool
	FailureReason string
}

func (m *MockWalletManagerForImport) ImportWallet(ctx context.Context, mnemonic, password, chainName, derivationPath string) (address string, publicKey string, importedAt int64, err error) {
	if m.ShouldFail {
		return "", "", 0, &ImportError{Reason: m.FailureReason}
	}
	return "0x1234567890abcdef1234567890abcdef12345678", "0x04abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", 1234567890, nil
}

// ImportError helps simulate specific error types
type ImportError struct {
	Reason string
}

func (e *ImportError) Error() string {
	return e.Reason
}

func TestCreateImportWalletHandler_Success(t *testing.T) {
	mockWalletManager := &wallet.MockWalletManager{}
	
	// Set up expectations for the mock
	mockWalletManager.On("ImportWallet", mock.Anything, "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about", "password123", "ethereum", "").Return(
		"0x1234567890abcdef1234567890abcdef12345678",
		"0x04abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		int64(1234567890),
		nil,
	)
	
	handler := CreateImportWalletHandler(mockWalletManager)

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
	
	// Assert that the expectations were met
	mockWalletManager.AssertExpectations(t)
}

func TestCreateImportWalletHandler_MissingParameters(t *testing.T) {
	mockWalletManager := &wallet.MockWalletManager{}
	mockManager := &MockWalletManagerForImport{
		MockWalletManager: mockWalletManager,
	}
	handler := CreateImportWalletHandler(mockManager)

	tests := []struct {
		name          string
		params        ImportWalletParams
		expectedError int
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
	mockWalletManager := &wallet.MockWalletManager{}
	mockManager := &MockWalletManagerForImport{
		MockWalletManager: mockWalletManager,
	}
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
		name          string
		failureReason string
		expectedError int
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
			mockWalletManager := &wallet.MockWalletManager{}
			mockManager := &MockWalletManagerForImport{
				MockWalletManager: mockWalletManager,
				ShouldFail:        true,
				FailureReason:     tt.failureReason,
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
	mockWalletManager := &wallet.MockWalletManager{}
	mockManager := &MockWalletManagerForImport{
		MockWalletManager: mockWalletManager,
	}
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
