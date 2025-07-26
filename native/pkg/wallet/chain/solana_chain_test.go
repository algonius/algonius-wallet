package chain

import (
	"context"
	"testing"
)

func TestSolanaChain_CreateWallet(t *testing.T) {
	chain := NewSolanaChain()
	ctx := context.Background()

	walletInfo, err := chain.CreateWallet(ctx)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if walletInfo == nil {
		t.Error("Expected wallet info, got nil")
	}

	if walletInfo.Address == "" {
		t.Error("Expected address to be set")
	}

	if walletInfo.PublicKey == "" {
		t.Error("Expected public key to be set")
	}

	if walletInfo.PrivateKey == "" {
		t.Error("Expected private key to be set")
	}

	if walletInfo.Mnemonic == "" {
		t.Error("Expected mnemonic to be set")
	}
}

func TestSolanaChain_GetBalance(t *testing.T) {
	chain := NewSolanaChain()
	ctx := context.Background()

	// Test valid address and SOL token
	address := "29VrjJ8FS898AjuSzi745x8971sF5Zu1pPfY4j6Fm4iA"
	balance, err := chain.GetBalance(ctx, address, "SOL")
	if err != nil {
		t.Errorf("Expected no error for valid address and SOL token, got: %v", err)
	}
	if balance != "0" {
		t.Errorf("Expected balance '0', got: %s", balance)
	}

	// Test valid address with empty token (should default to SOL)
	balance, err = chain.GetBalance(ctx, address, "")
	if err != nil {
		t.Errorf("Expected no error for valid address with empty token, got: %v", err)
	}
	if balance != "0" {
		t.Errorf("Expected balance '0', got: %s", balance)
	}

	// Test unsupported token
	_, err = chain.GetBalance(ctx, address, "USDC")
	if err == nil {
		t.Error("Expected error for unsupported token")
	}
}

func TestSolanaChain_GetChainName(t *testing.T) {
	chain := NewSolanaChain()
	if chain.GetChainName() != "SOLANA" {
		t.Errorf("Expected chain name 'SOLANA', got: %s", chain.GetChainName())
	}
}

func TestChainFactory_SolanaChain(t *testing.T) {
	factory := NewChainFactory()

	// Test SOL chain name
	chain, err := factory.GetChain("SOL")
	if err != nil {
		t.Errorf("Expected no error for SOL chain, got: %v", err)
	}
	if chain == nil {
		t.Error("Expected chain instance for SOL, got nil")
	}
	if chain.GetChainName() != "SOLANA" {
		t.Errorf("Expected chain name 'SOLANA', got: %s", chain.GetChainName())
	}

	// Test SOLANA chain name
	chain, err = factory.GetChain("SOLANA")
	if err != nil {
		t.Errorf("Expected no error for SOLANA chain, got: %v", err)
	}
	if chain == nil {
		t.Error("Expected chain instance for SOLANA, got nil")
	}
	if chain.GetChainName() != "SOLANA" {
		t.Errorf("Expected chain name 'SOLANA', got: %s", chain.GetChainName())
	}
}