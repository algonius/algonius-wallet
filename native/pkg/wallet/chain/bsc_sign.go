// SPDX-License-Identifier: Apache-2.0
package chain

// SignMessage signs a message using the provided private key
// For BSC, we use the same signing method as Ethereum (EIP-191)
func (b *BSCChain) SignMessage(privateKeyHex, message string) (string, error) {
	// BSC uses the same signing method as Ethereum
	// We can reuse the Ethereum implementation by creating a temporary ETHChain instance
	ethChain := NewETHChainLegacy()
	return ethChain.SignMessage(privateKeyHex, message)
}