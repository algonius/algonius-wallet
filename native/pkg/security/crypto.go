// SPDX-License-Identifier: Apache-2.0
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// AESKeySize is the size of AES keys in bytes
	AESKeySize = 32 // AES-256

	// SaltSize is the size of the salt in bytes
	SaltSize = 32

	// PBKDF2Iterations is the number of iterations for PBKDF2
	PBKDF2Iterations = 100000
)

// EncryptedData represents encrypted data with metadata
type EncryptedData struct {
	Data string `json:"data"` // Base64 encoded encrypted data
	Salt string `json:"salt"` // Base64 encoded salt
}

// EncryptWithPassword encrypts data using a password
func EncryptWithPassword(data, password string) (*EncryptedData, error) {
	if data == "" {
		return nil, errors.New("data cannot be empty")
	}
	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	// Generate random salt
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from password using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, AESKeySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)

	// Encode to base64
	encryptedData := base64.StdEncoding.EncodeToString(ciphertext)
	saltEncoded := base64.StdEncoding.EncodeToString(salt)

	return &EncryptedData{
		Data: encryptedData,
		Salt: saltEncoded,
	}, nil
}

// DecryptWithPassword decrypts data using a password
func DecryptWithPassword(encryptedData *EncryptedData, password string) (string, error) {
	if encryptedData == nil {
		return "", errors.New("encrypted data cannot be nil")
	}
	if encryptedData.Data == "" {
		return "", errors.New("encrypted data cannot be empty")
	}
	if encryptedData.Salt == "" {
		return "", errors.New("salt cannot be empty")
	}
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// Decode base64 data
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData.Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	// Decode salt
	salt, err := base64.StdEncoding.DecodeString(encryptedData.Salt)
	if err != nil {
		return "", fmt.Errorf("failed to decode salt: %w", err)
	}

	// Derive key from password using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, AESKeySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check minimum ciphertext length
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %w", err)
	}

	return string(plaintext), nil
}
