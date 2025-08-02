// SPDX-License-Identifier: Apache-2.0
package chain

import (
	"fmt"

	"github.com/mr-tron/base58"
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

	// Convert message to bytes if it's a hex string
	var messageBytes []byte
	if len(message) > 2 && message[0:2] == "0x" {
		// If it's a hex string, decode it
		decoded, err := base58.Decode(message[2:]) // Skip "0x" prefix
		if err != nil {
			return "", fmt.Errorf("failed to decode hex message: %w", err)
		}
		messageBytes = decoded
	} else {
		// Otherwise, treat it as a regular string
		messageBytes = []byte(message)
	}

	// Sign the message using Ed25519
	signature := ed25519.Sign(privateKey, messageBytes)

	// For Solana, we return the raw 64-byte signature as base58
	signatureBase58 := base58.Encode(signature)
	return signatureBase58, nil
}