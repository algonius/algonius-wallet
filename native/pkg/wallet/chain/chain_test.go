package chain

import (
	"context"
	"testing"
)

func TestETHChain_ConfirmTransaction(t *testing.T) {
	chain := NewETHChain()
	ctx := context.Background()

	// Test valid transaction hash (64 chars + 0x = 66 total)
	txHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	confirmation, err := chain.ConfirmTransaction(ctx, txHash, 6)
	
	if err != nil {
		t.Errorf("Expected no error for valid transaction hash, got: %v", err)
	}

	if confirmation == nil {
		t.Error("Expected confirmation result, got nil")
	}

	if confirmation.RequiredConfirmations != 6 {
		t.Errorf("Expected required confirmations to be 6, got %d", confirmation.RequiredConfirmations)
	}

	if confirmation.TxHash != txHash {
		t.Errorf("Expected tx hash to be %s, got %s", txHash, confirmation.TxHash)
	}

	// Verify status is one of the valid values
	validStatuses := []string{"pending", "confirmed", "failed"}
	found := false
	for _, status := range validStatuses {
		if confirmation.Status == status {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected status to be one of %v, got %s", validStatuses, confirmation.Status)
	}
}

func TestETHChain_ConfirmTransaction_DefaultConfirmations(t *testing.T) {
	chain := NewETHChain()
	ctx := context.Background()

	txHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	
	// Test with zero required confirmations (should default to 6 for ETH)
	confirmation, err := chain.ConfirmTransaction(ctx, txHash, 0)
	
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if confirmation.RequiredConfirmations != 6 {
		t.Errorf("Expected default ETH confirmations to be 6, got %d", confirmation.RequiredConfirmations)
	}
}

func TestETHChain_ConfirmTransaction_InvalidHash(t *testing.T) {
	chain := NewETHChain()
	ctx := context.Background()

	// Test empty hash
	_, err := chain.ConfirmTransaction(ctx, "", 6)
	if err == nil {
		t.Error("Expected error for empty transaction hash")
	}

	// Test invalid length
	_, err = chain.ConfirmTransaction(ctx, "0x123", 6)
	if err == nil {
		t.Error("Expected error for invalid transaction hash length")
	}

	// Test invalid hex characters
	_, err = chain.ConfirmTransaction(ctx, "0xgggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg", 6)
	if err == nil {
		t.Error("Expected error for invalid hex characters")
	}
}

func TestETHChain_ConfirmTransaction_HashNormalization(t *testing.T) {
	chain := NewETHChain()
	ctx := context.Background()

	// Test hash without 0x prefix
	hashWithoutPrefix := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	confirmation, err := chain.ConfirmTransaction(ctx, hashWithoutPrefix, 6)
	
	if err != nil {
		t.Errorf("Expected no error for hash without prefix, got: %v", err)
	}

	expectedHash := "0x" + hashWithoutPrefix
	if confirmation.TxHash != expectedHash {
		t.Errorf("Expected normalized hash %s, got %s", expectedHash, confirmation.TxHash)
	}
}

func TestBSCChain_ConfirmTransaction(t *testing.T) {
	chain := NewBSCChain()
	ctx := context.Background()

	// Test valid transaction hash
	txHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	confirmation, err := chain.ConfirmTransaction(ctx, txHash, 3)
	
	if err != nil {
		t.Errorf("Expected no error for valid transaction hash, got: %v", err)
	}

	if confirmation == nil {
		t.Error("Expected confirmation result, got nil")
	}

	if confirmation.RequiredConfirmations != 3 {
		t.Errorf("Expected required confirmations to be 3, got %d", confirmation.RequiredConfirmations)
	}

	// BSC should have higher block numbers than ETH
	if confirmation.BlockNumber < 30000000 {
		t.Errorf("Expected BSC block number to be higher, got %d", confirmation.BlockNumber)
	}
}

func TestBSCChain_ConfirmTransaction_DefaultConfirmations(t *testing.T) {
	chain := NewBSCChain()
	ctx := context.Background()

	txHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	
	// Test with zero required confirmations (should default to 3 for BSC)
	confirmation, err := chain.ConfirmTransaction(ctx, txHash, 0)
	
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if confirmation.RequiredConfirmations != 3 {
		t.Errorf("Expected default BSC confirmations to be 3, got %d", confirmation.RequiredConfirmations)
	}
}

func TestChainFactory_GetChain(t *testing.T) {
	factory := NewChainFactory()

	// Test valid chains
	validChains := []string{"ETH", "ETHEREUM", "BSC", "BINANCE"}
	for _, chainName := range validChains {
		chain, err := factory.GetChain(chainName)
		if err != nil {
			t.Errorf("Expected no error for chain %s, got: %v", chainName, err)
		}
		if chain == nil {
			t.Errorf("Expected chain instance for %s, got nil", chainName)
		}
	}

	// Test invalid chain
	_, err := factory.GetChain("INVALID")
	if err == nil {
		t.Error("Expected error for invalid chain name")
	}
}

func TestChainFactory_GetSupportedChains(t *testing.T) {
	factory := NewChainFactory()
	
	chains := factory.GetSupportedChains()
	
	if len(chains) == 0 {
		t.Error("Expected at least one supported chain")
	}

	// Should contain ETH and BSC
	hasETH := false
	hasBSC := false
	for _, chain := range chains {
		if chain == "ETH" {
			hasETH = true
		}
		if chain == "BSC" {
			hasBSC = true
		}
	}

	if !hasETH {
		t.Error("Expected ETH to be in supported chains")
	}
	if !hasBSC {
		t.Error("Expected BSC to be in supported chains")
	}
}