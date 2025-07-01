// SPDX-License-Identifier: Apache-2.0
package security

import (
	"testing"
)

func TestEncryptDecryptWithPassword(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		password string
	}{
		{
			name:     "simple text",
			data:     "hello world",
			password: "password123",
		},
		{
			name:     "private key",
			data:     "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			password: "secure_password_456",
		},
		{
			name:     "mnemonic phrase",
			data:     "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			password: "very_secure_password_789",
		},
		{
			name:     "special characters",
			data:     "Test data with special chars: !@#$%^&*()_+{}|:<>?[]\\;'\",./-",
			password: "password_with_special_chars!@#",
		},
		{
			name:     "unicode data",
			data:     "Unicode test: 你好世界 🌍 ñañá ëxämplê",
			password: "unicode_password_测试",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test encryption
			encrypted, err := EncryptWithPassword(tt.data, tt.password)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Verify encrypted data is not empty
			if encrypted.Data == "" {
				t.Error("Encrypted data is empty")
			}
			if encrypted.Salt == "" {
				t.Error("Salt is empty")
			}

			// Verify encrypted data is different from original
			if encrypted.Data == tt.data {
				t.Error("Encrypted data is the same as original data")
			}

			// Test decryption
			decrypted, err := DecryptWithPassword(encrypted, tt.password)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			// Verify decrypted data matches original
			if decrypted != tt.data {
				t.Errorf("Decrypted data '%s' does not match original '%s'", decrypted, tt.data)
			}
		})
	}
}

func TestEncryptWithPassword_ErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		password string
		expectErr bool
	}{
		{
			name:      "empty data",
			data:      "",
			password:  "password123",
			expectErr: true,
		},
		{
			name:      "empty password",
			data:      "test data",
			password:  "",
			expectErr: true,
		},
		{
			name:      "both empty",
			data:      "",
			password:  "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EncryptWithPassword(tt.data, tt.password)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestDecryptWithPassword_ErrorCases(t *testing.T) {
	// First create valid encrypted data for testing
	validData := "test data"
	validPassword := "password123"
	validEncrypted, err := EncryptWithPassword(validData, validPassword)
	if err != nil {
		t.Fatalf("Failed to create valid encrypted data for testing: %v", err)
	}

	tests := []struct {
		name      string
		encrypted *EncryptedData
		password  string
		expectErr bool
	}{
		{
			name:      "nil encrypted data",
			encrypted: nil,
			password:  "password123",
			expectErr: true,
		},
		{
			name: "empty encrypted data",
			encrypted: &EncryptedData{
				Data: "",
				Salt: validEncrypted.Salt,
			},
			password:  "password123",
			expectErr: true,
		},
		{
			name: "empty salt",
			encrypted: &EncryptedData{
				Data: validEncrypted.Data,
				Salt: "",
			},
			password:  "password123",
			expectErr: true,
		},
		{
			name:      "empty password",
			encrypted: validEncrypted,
			password:  "",
			expectErr: true,
		},
		{
			name:      "wrong password",
			encrypted: validEncrypted,
			password:  "wrong_password",
			expectErr: true,
		},
		{
			name: "invalid base64 data",
			encrypted: &EncryptedData{
				Data: "invalid_base64_!@#$%",
				Salt: validEncrypted.Salt,
			},
			password:  "password123",
			expectErr: true,
		},
		{
			name: "invalid base64 salt",
			encrypted: &EncryptedData{
				Data: validEncrypted.Data,
				Salt: "invalid_base64_!@#$%",
			},
			password:  "password123",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptWithPassword(tt.encrypted, tt.password)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestEncryption_DifferentResults(t *testing.T) {
	// Test that encrypting the same data twice produces different results
	data := "test data"
	password := "password123"

	encrypted1, err := EncryptWithPassword(data, password)
	if err != nil {
		t.Fatalf("First encryption failed: %v", err)
	}

	encrypted2, err := EncryptWithPassword(data, password)
	if err != nil {
		t.Fatalf("Second encryption failed: %v", err)
	}

	// Encrypted data should be different due to random nonce/salt
	if encrypted1.Data == encrypted2.Data {
		t.Error("Encrypted data should be different for each encryption")
	}

	if encrypted1.Salt == encrypted2.Salt {
		t.Error("Salt should be different for each encryption")
	}

	// But both should decrypt to the same original data
	decrypted1, err := DecryptWithPassword(encrypted1, password)
	if err != nil {
		t.Fatalf("First decryption failed: %v", err)
	}

	decrypted2, err := DecryptWithPassword(encrypted2, password)
	if err != nil {
		t.Fatalf("Second decryption failed: %v", err)
	}

	if decrypted1 != data || decrypted2 != data {
		t.Error("Both decryptions should produce the original data")
	}
}