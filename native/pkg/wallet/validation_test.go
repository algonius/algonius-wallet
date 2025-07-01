// SPDX-License-Identifier: Apache-2.0
package wallet

import (
	"testing"
)

func TestValidateMnemonic(t *testing.T) {
	tests := []struct {
		name      string
		mnemonic  string
		expectErr bool
	}{
		{
			name:      "valid 12-word mnemonic",
			mnemonic:  "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			expectErr: false,
		},
		{
			name:      "valid 24-word mnemonic",
			mnemonic:  "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
			expectErr: false,
		},
		{
			name:      "empty mnemonic",
			mnemonic:  "",
			expectErr: true,
		},
		{
			name:      "invalid word count (11 words)",
			mnemonic:  "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon",
			expectErr: true,
		},
		{
			name:      "invalid word count (13 words)",
			mnemonic:  "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about extra",
			expectErr: true,
		},
		{
			name:      "invalid checksum",
			mnemonic:  "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon",
			expectErr: true,
		},
		{
			name:      "mnemonic with extra spaces",
			mnemonic:  "  abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about  ",
			expectErr: false,
		},
		{
			name:      "mnemonic with mixed case",
			mnemonic:  "Abandon Abandon abandon ABANDON abandon abandon abandon abandon abandon abandon abandon about",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMnemonic(tt.mnemonic)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error for mnemonic '%s', but got none", tt.mnemonic)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for mnemonic '%s', but got: %v", tt.mnemonic, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		expectErr bool
	}{
		{
			name:      "valid password",
			password:  "password123",
			expectErr: false,
		},
		{
			name:      "long valid password",
			password:  "this_is_a_very_long_and_secure_password_123!@#",
			expectErr: false,
		},
		{
			name:      "empty password",
			password:  "",
			expectErr: true,
		},
		{
			name:      "short password",
			password:  "pass",
			expectErr: true,
		},
		{
			name:      "7 character password",
			password:  "1234567",
			expectErr: true,
		},
		{
			name:      "8 character password",
			password:  "12345678",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error for password '%s', but got none", tt.password)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for password '%s', but got: %v", tt.password, err)
			}
		})
	}
}

func TestValidateChain(t *testing.T) {
	tests := []struct {
		name      string
		chain     string
		expectErr bool
	}{
		{
			name:      "ethereum",
			chain:     "ethereum",
			expectErr: false,
		},
		{
			name:      "eth",
			chain:     "eth",
			expectErr: false,
		},
		{
			name:      "ETH uppercase",
			chain:     "ETH",
			expectErr: false,
		},
		{
			name:      "bsc",
			chain:     "bsc",
			expectErr: false,
		},
		{
			name:      "BSC uppercase",
			chain:     "BSC",
			expectErr: false,
		},
		{
			name:      "binance",
			chain:     "binance",
			expectErr: false,
		},
		{
			name:      "empty chain",
			chain:     "",
			expectErr: true,
		},
		{
			name:      "unsupported chain",
			chain:     "bitcoin",
			expectErr: true,
		},
		{
			name:      "chain with spaces",
			chain:     "  ethereum  ",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChain(tt.chain)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error for chain '%s', but got none", tt.chain)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for chain '%s', but got: %v", tt.chain, err)
			}
		})
	}
}

func TestNormalizeChain(t *testing.T) {
	tests := []struct {
		name     string
		chain    string
		expected string
	}{
		{
			name:     "ethereum",
			chain:    "ethereum",
			expected: "ethereum",
		},
		{
			name:     "eth",
			chain:    "eth",
			expected: "ethereum",
		},
		{
			name:     "ETH uppercase",
			chain:    "ETH",
			expected: "ethereum",
		},
		{
			name:     "Ethereum mixed case",
			chain:    "Ethereum",
			expected: "ethereum",
		},
		{
			name:     "bsc",
			chain:    "bsc",
			expected: "bsc",
		},
		{
			name:     "BSC uppercase",
			chain:    "BSC",
			expected: "bsc",
		},
		{
			name:     "binance",
			chain:    "binance",
			expected: "bsc",
		},
		{
			name:     "chain with spaces",
			chain:    "  ethereum  ",
			expected: "ethereum",
		},
		{
			name:     "unsupported chain",
			chain:    "bitcoin",
			expected: "bitcoin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeChain(tt.chain)
			if result != tt.expected {
				t.Errorf("Expected '%s' for chain '%s', got '%s'", tt.expected, tt.chain, result)
			}
		})
	}
}