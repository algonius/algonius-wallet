// SPDX-License-Identifier: Apache-2.0
package chain

import (
	"fmt"
	"strings"

	"github.com/mr-tron/base58"
	"go.uber.org/zap"
	"golang.org/x/crypto/ed25519"
)

// SignMessage signs a message using the provided private key
// For Solana, we use Ed25519 signing which produces a 64-byte signature
func (s *SolanaChain) SignMessage(privateKeyHex, message string) (string, error) {
	// Decode the private key from base58 format (Solana uses base58 for keys)
	privateKeyBytes, err := base58.Decode(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	// Validate private key length (Ed25519 private key should be 64 bytes)
	// The first 32 bytes are the seed, and the full 64 bytes include the public key
	if len(privateKeyBytes) != 64 {
		return "", fmt.Errorf("invalid private key length: expected 64 bytes, got %d", len(privateKeyBytes))
	}

	// Convert to ed25519.PrivateKey
	privateKey := ed25519.PrivateKey(privateKeyBytes)

	// Check if this is a special Solana raw bytes message
	var messageBytes []byte
	if strings.HasPrefix(message, "__SOLANA_RAW_BYTES__:") {
		// Extract the raw bytes after the marker
		messageBytes = []byte(message[21:]) // Skip "__SOLANA_RAW_BYTES__:"
	} else {
		// Otherwise, treat it as a regular string
		messageBytes = []byte(message)
	}

	// Sign the message using Ed25519 (64-byte signature, no v value)
	signature := ed25519.Sign(privateKey, messageBytes)

	// Ensure the signature is exactly 64 bytes
	if len(signature) != 64 {
		return "", fmt.Errorf("invalid signature length: expected 64 bytes, got %d", len(signature))
	}

	// For Solana, we return the raw 64-byte signature as base58
	signatureBase58 := base58.Encode(signature)
	
	// Log the signature for debugging
	s.logger.Debug("Solana signature generated", 
		zap.String("signature_base58", signatureBase58),
		zap.Int("signature_length", len(signature)))
	
	return signatureBase58, nil
}