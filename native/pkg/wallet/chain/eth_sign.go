// SPDX-License-Identifier: Apache-2.0
package chain

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// SignMessage signs a message using the provided private key
// The message is first hashed using Ethereum's signing standard (EIP-191)
// which prefixes the message with "\x19Ethereum Signed Message:\n" + len(message)
func (e *ETHChain) SignMessage(privateKeyHex, message string) (string, error) {
	// Parse the private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Convert message to bytes if it's a hex string
	var messageBytes []byte
	if strings.HasPrefix(message, "0x") {
		// If it's a hex string, decode it
		messageBytes, err = hexutil.Decode(message)
		if err != nil {
			return "", fmt.Errorf("failed to decode hex message: %w", err)
		}
	} else {
		// Otherwise, treat it as a regular string
		messageBytes = []byte(message)
	}

	// Sign the message using Ethereum's signing standard (EIP-191)
	signature, err := crypto.Sign(crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(messageBytes), messageBytes))).Bytes(), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	// Convert the signature to hex format
	signatureHex := hexutil.Encode(signature)
	return signatureHex, nil
}