// SPDX-License-Identifier: Apache-2.0
package wallet

import (
	"errors"
	"fmt"
	"strings"

	bip39 "github.com/tyler-smith/go-bip39"
)

// ValidateMnemonic validates a BIP39 mnemonic phrase
func ValidateMnemonic(mnemonic string) error {
	if mnemonic == "" {
		return errors.New("mnemonic phrase cannot be empty")
	}

	// Normalize the mnemonic (trim spaces, lowercase)
	normalizedMnemonic := strings.TrimSpace(strings.ToLower(mnemonic))
	
	// Check word count (BIP39 supports 12, 15, 18, 21, 24 words)
	words := strings.Fields(normalizedMnemonic)
	wordCount := len(words)
	
	validWordCounts := []int{12, 15, 18, 21, 24}
	isValidCount := false
	for _, count := range validWordCounts {
		if wordCount == count {
			isValidCount = true
			break
		}
	}
	
	if !isValidCount {
		return fmt.Errorf("invalid mnemonic word count: %d (must be 12, 15, 18, 21, or 24 words)", wordCount)
	}

	// Check for duplicate words (common issue that makes BIP39 validation fail)
	wordMap := make(map[string][]int)
	for i, word := range words {
		wordMap[word] = append(wordMap[word], i+1) // 1-based position for user-friendly error
	}
	
	var duplicates []string
	for word, positions := range wordMap {
		if len(positions) > 1 {
			duplicates = append(duplicates, fmt.Sprintf("'%s' (positions %v)", word, positions))
		}
	}
	
	if len(duplicates) > 0 {
		return fmt.Errorf("invalid mnemonic: duplicate words found: %s", strings.Join(duplicates, ", "))
	}
	
	// Check if the mnemonic is valid according to BIP39
	if !bip39.IsMnemonicValid(normalizedMnemonic) {
		return errors.New("invalid mnemonic phrase")
	}

	return nil
}

// ValidatePassword performs basic password validation
func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Additional password strength checks could be added here
	// For now, we keep it simple but functional

	return nil
}

// ValidateChain validates if the chain is supported for wallet import
func ValidateChain(chain string) error {
	if chain == "" {
		return errors.New("chain cannot be empty")
	}

	// Normalize chain name
	normalizedChain := strings.ToLower(strings.TrimSpace(chain))
	
	supportedChains := []string{"ethereum", "eth", "bsc", "binance"}
	for _, supported := range supportedChains {
		if normalizedChain == supported {
			return nil
		}
	}

	return fmt.Errorf("unsupported chain: %s (supported: ethereum, bsc)", chain)
}

// NormalizeChain normalizes chain names to standard format
func NormalizeChain(chain string) string {
	normalizedChain := strings.ToLower(strings.TrimSpace(chain))
	
	switch normalizedChain {
	case "eth", "ethereum":
		return "ethereum"
	case "bsc", "binance":
		return "bsc"
	default:
		return normalizedChain
	}
}