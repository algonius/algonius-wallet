package wallet

import (
	"context"
	"testing"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/wallet/chain"
)

func TestTransactionCache_SetAndGet(t *testing.T) {
	cache := NewTransactionCache(1 * time.Second)

	confirmation := &chain.TransactionConfirmation{
		Status:        "confirmed",
		Confirmations: 10,
		TxHash:        "0x123",
	}

	// Test Set and immediate Get
	cache.Set("test-key", confirmation)
	
	retrieved, found := cache.Get("test-key")
	if !found {
		t.Error("Expected to find cached confirmation")
	}
	
	if retrieved.Status != "confirmed" {
		t.Errorf("Expected status 'confirmed', got '%s'", retrieved.Status)
	}
}

func TestTransactionCache_Expiry(t *testing.T) {
	cache := NewTransactionCache(100 * time.Millisecond) // Very short TTL

	confirmation := &chain.TransactionConfirmation{
		Status: "confirmed",
		TxHash: "0x123",
	}

	cache.Set("test-key", confirmation)
	
	// Should be found immediately
	_, found := cache.Get("test-key")
	if !found {
		t.Error("Expected to find cached confirmation immediately")
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)
	
	// Should be expired now
	_, found = cache.Get("test-key")
	if found {
		t.Error("Expected confirmation to be expired")
	}
}

func TestTransactionCache_CleanupExpired(t *testing.T) {
	cache := NewTransactionCache(100 * time.Millisecond)

	confirmation := &chain.TransactionConfirmation{
		Status: "confirmed",
		TxHash: "0x123",
	}

	cache.Set("test-key", confirmation)
	
	// Wait for expiry
	time.Sleep(150 * time.Millisecond)
	
	// Cleanup expired entries
	cache.CleanupExpired()
	
	// Check that the entry is gone from the internal cache
	cache.mutex.RLock()
	_, exists := cache.cache["test-key"]
	cache.mutex.RUnlock()
	
	if exists {
		t.Error("Expected expired entry to be cleaned up")
	}
}

func TestConfirmTransaction_EmptyParams(t *testing.T) {
	factory := chain.NewChainFactory()
	ctx := context.Background()

	// Test empty chain
	_, err := ConfirmTransaction(ctx, "", "0x123", 6, factory)
	if err == nil {
		t.Error("Expected error for empty chain name")
	}

	// Test empty tx hash
	_, err = ConfirmTransaction(ctx, "ETH", "", 6, factory)
	if err == nil {
		t.Error("Expected error for empty transaction hash")
	}
}

func TestConfirmTransaction_UnsupportedChain(t *testing.T) {
	factory := chain.NewChainFactory()
	ctx := context.Background()

	_, err := ConfirmTransaction(ctx, "UNSUPPORTED", "0x123", 6, factory)
	if err == nil {
		t.Error("Expected error for unsupported chain")
	}
}

func TestConfirmTransaction_ValidRequest(t *testing.T) {
	factory := chain.NewChainFactory()
	ctx := context.Background()

	// Test with valid ETH transaction
	confirmation, err := ConfirmTransaction(ctx, "ETH", "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", 6, factory)
	if err != nil {
		t.Errorf("Expected no error for valid ETH transaction, got: %v", err)
	}

	if confirmation == nil {
		t.Error("Expected confirmation result, got nil")
	}

	if confirmation.RequiredConfirmations != 6 {
		t.Errorf("Expected required confirmations to be 6, got %d", confirmation.RequiredConfirmations)
	}

	// Test with valid BSC transaction
	confirmation, err = ConfirmTransaction(ctx, "BSC", "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", 3, factory)
	if err != nil {
		t.Errorf("Expected no error for valid BSC transaction, got: %v", err)
	}

	if confirmation == nil {
		t.Error("Expected confirmation result, got nil")
	}

	if confirmation.RequiredConfirmations != 3 {
		t.Errorf("Expected required confirmations to be 3, got %d", confirmation.RequiredConfirmations)
	}
}

func TestConfirmTransaction_DefaultConfirmations(t *testing.T) {
	factory := chain.NewChainFactory()
	ctx := context.Background()

	// Test with zero required confirmations (should use defaults)
	confirmation, err := ConfirmTransaction(ctx, "ETH", "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", 0, factory)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// ETH default should be 6
	if confirmation.RequiredConfirmations != 6 {
		t.Errorf("Expected default ETH confirmations to be 6, got %d", confirmation.RequiredConfirmations)
	}

	// Test BSC default
	confirmation, err = ConfirmTransaction(ctx, "BSC", "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", 0, factory)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// BSC default should be 3
	if confirmation.RequiredConfirmations != 3 {
		t.Errorf("Expected default BSC confirmations to be 3, got %d", confirmation.RequiredConfirmations)
	}
}