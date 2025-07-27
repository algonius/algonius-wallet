// SPDX-License-Identifier: Apache-2.0
package dex

import (
	"context"
	"testing"

	"go.uber.org/zap/zaptest"
)

// MockProvider implements IDEXProvider for testing
type MockProvider struct {
	name            string
	supportedChains map[string]bool
	shouldFail      bool
	quoteResponse   *SwapQuote
	swapResponse    *SwapResult
	balanceResponse *BalanceInfo
}

func NewMockProvider(name string, supportedChains []string) *MockProvider {
	chains := make(map[string]bool)
	for _, chain := range supportedChains {
		chains[chain] = true
	}
	
	return &MockProvider{
		name:            name,
		supportedChains: chains,
		shouldFail:      false,
		quoteResponse: &SwapQuote{
			FromToken:    "ETH",
			ToToken:      "USDT",
			FromAmount:   "1.0",
			ToAmount:     "3000.0",
			EstimatedGas: 150000,
			GasPrice:     "20",
			EstimatedFee: "0.003",
			Slippage:     0.005,
			PriceImpact:  0.001,
			Route:        []string{"Uniswap V2"},
			ValidUntil:   9999999999,
			Provider:     name,
		},
		swapResponse: &SwapResult{
			TxHash:     "0xabcdef1234567890",
			Status:     "pending",
			FromToken:  "ETH",
			ToToken:    "USDT",
			FromAmount: "1.0",
			ToAmount:   "3000.0",
			ActualFee:  "0.003",
			Provider:   name,
			Timestamp:  1700000000,
		},
		balanceResponse: &BalanceInfo{
			TokenAddress: "0xA0b86a33E6417aAA9f79dF12E43d04Be87da7C9c",
			TokenSymbol:  "USDC",
			Balance:      "1000.000000",
			Decimals:     6,
			USDValue:     "1000.00",
		},
	}
}

func (m *MockProvider) GetName() string {
	return m.name
}

func (m *MockProvider) IsSupported(chainID string) bool {
	return m.supportedChains[chainID]
}

func (m *MockProvider) GetQuote(ctx context.Context, params SwapParams) (*SwapQuote, error) {
	if m.shouldFail {
		return nil, &ValidationError{Field: "mock", Message: "mock provider failure"}
	}
	
	quote := *m.quoteResponse // Copy
	quote.FromToken = params.FromToken
	quote.ToToken = params.ToToken
	quote.FromAmount = params.Amount
	quote.Slippage = params.Slippage
	return &quote, nil
}

func (m *MockProvider) ExecuteSwap(ctx context.Context, params SwapParams) (*SwapResult, error) {
	if m.shouldFail {
		return nil, &ValidationError{Field: "mock", Message: "mock provider failure"}
	}
	
	result := *m.swapResponse // Copy
	result.FromToken = params.FromToken
	result.ToToken = params.ToToken
	result.FromAmount = params.Amount
	return &result, nil
}

func (m *MockProvider) GetBalance(ctx context.Context, address string, tokenAddress string, chainID string) (*BalanceInfo, error) {
	if m.shouldFail {
		return nil, &ValidationError{Field: "mock", Message: "mock provider failure"}
	}
	return m.balanceResponse, nil
}

func (m *MockProvider) EstimateGas(ctx context.Context, params SwapParams) (gasLimit uint64, gasPrice string, err error) {
	if m.shouldFail {
		return 0, "", &ValidationError{Field: "mock", Message: "mock provider failure"}
	}
	return 150000, "20", nil
}

func (m *MockProvider) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

func TestDEXAggregator_RegisterProvider(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aggregator := NewDEXAggregator(logger)
	
	mockProvider := NewMockProvider("MockDEX", []string{"1", "56"})
	
	err := aggregator.RegisterProvider(mockProvider)
	if err != nil {
		t.Errorf("Failed to register provider: %v", err)
	}
	
	// Try to register the same provider again
	err = aggregator.RegisterProvider(mockProvider)
	if err == nil {
		t.Error("Expected error when registering duplicate provider")
	}
}

func TestDEXAggregator_GetSupportedProviders(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aggregator := NewDEXAggregator(logger)
	
	// Register providers with different chain support
	provider1 := NewMockProvider("Provider1", []string{"1"})    // Ethereum only
	provider2 := NewMockProvider("Provider2", []string{"56"})   // BSC only
	provider3 := NewMockProvider("Provider3", []string{"1", "56"}) // Both
	
	aggregator.RegisterProvider(provider1)
	aggregator.RegisterProvider(provider2)
	aggregator.RegisterProvider(provider3)
	
	// Test Ethereum support
	ethProviders := aggregator.GetSupportedProviders("1")
	expectedEthProviders := 2 // Provider1 and Provider3
	if len(ethProviders) != expectedEthProviders {
		t.Errorf("Expected %d providers for Ethereum, got %d", expectedEthProviders, len(ethProviders))
	}
	
	// Test BSC support
	bscProviders := aggregator.GetSupportedProviders("56")
	expectedBscProviders := 2 // Provider2 and Provider3
	if len(bscProviders) != expectedBscProviders {
		t.Errorf("Expected %d providers for BSC, got %d", expectedBscProviders, len(bscProviders))
	}
	
	// Test unsupported chain
	solProviders := aggregator.GetSupportedProviders("501")
	if len(solProviders) != 0 {
		t.Errorf("Expected 0 providers for Solana, got %d", len(solProviders))
	}
}

func TestDEXAggregator_GetBestQuote(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aggregator := NewDEXAggregator(logger)
	
	// Create providers with different quote amounts
	provider1 := NewMockProvider("Provider1", []string{"1"})
	provider1.quoteResponse.ToAmount = "3000.0" // Lower amount
	
	provider2 := NewMockProvider("Provider2", []string{"1"})
	provider2.quoteResponse.ToAmount = "3100.0" // Higher amount (better)
	
	aggregator.RegisterProvider(provider1)
	aggregator.RegisterProvider(provider2)
	
	ctx := context.Background()
	params := SwapParams{
		FromToken:   "ETH",
		ToToken:     "USDT",
		Amount:      "1.0",
		Slippage:    0.005,
		FromAddress: "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
		ToAddress:   "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
		ChainID:     "1",
		PrivateKey:  "0x1234567890abcdef",
	}
	
	quote, err := aggregator.GetBestQuote(ctx, params)
	if err != nil {
		t.Errorf("Failed to get best quote: %v", err)
	}
	
	if quote.Provider != "Provider2" {
		t.Errorf("Expected best quote from Provider2, got %s", quote.Provider)
	}
	
	if quote.ToAmount != "3100.0" {
		t.Errorf("Expected ToAmount 3100.0, got %s", quote.ToAmount)
	}
}

func TestDEXAggregator_GetBestQuote_NoProviders(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aggregator := NewDEXAggregator(logger)
	
	ctx := context.Background()
	params := SwapParams{
		FromToken:   "ETH",
		ToToken:     "USDT",
		Amount:      "1.0",
		Slippage:    0.005,
		FromAddress: "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
		ToAddress:   "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
		ChainID:     "1",
		PrivateKey:  "0x1234567890abcdef",
	}
	
	_, err := aggregator.GetBestQuote(ctx, params)
	if err == nil {
		t.Error("Expected error when no providers support the chain")
	}
}

func TestDEXAggregator_ExecuteSwapWithProvider(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aggregator := NewDEXAggregator(logger)
	
	provider := NewMockProvider("TestProvider", []string{"1"})
	aggregator.RegisterProvider(provider)
	
	ctx := context.Background()
	params := SwapParams{
		FromToken:   "ETH",
		ToToken:     "USDT",
		Amount:      "1.0",
		Slippage:    0.005,
		FromAddress: "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
		ToAddress:   "0x742d35Cc6673C4C5f9aB9e3Be0A78a19a4B43c89",
		ChainID:     "1",
		PrivateKey:  "0x1234567890abcdef",
	}
	
	result, err := aggregator.ExecuteSwapWithProvider(ctx, "TestProvider", params)
	if err != nil {
		t.Errorf("Failed to execute swap: %v", err)
	}
	
	if result.Provider != "TestProvider" {
		t.Errorf("Expected provider TestProvider, got %s", result.Provider)
	}
	
	if result.Status != "pending" {
		t.Errorf("Expected status pending, got %s", result.Status)
	}
}

func TestDEXAggregator_ExecuteSwapWithProvider_NonExistentProvider(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aggregator := NewDEXAggregator(logger)
	
	ctx := context.Background()
	params := SwapParams{
		ChainID: "1",
	}
	
	_, err := aggregator.ExecuteSwapWithProvider(ctx, "NonExistentProvider", params)
	if err == nil {
		t.Error("Expected error when using non-existent provider")
	}
}

func TestDEXAggregator_GetProviderByName(t *testing.T) {
	logger := zaptest.NewLogger(t)
	aggregator := NewDEXAggregator(logger)
	
	provider := NewMockProvider("TestProvider", []string{"1"})
	aggregator.RegisterProvider(provider)
	
	retrievedProvider, err := aggregator.GetProviderByName("TestProvider")
	if err != nil {
		t.Errorf("Failed to get provider by name: %v", err)
	}
	
	if retrievedProvider.GetName() != "TestProvider" {
		t.Errorf("Expected provider name TestProvider, got %s", retrievedProvider.GetName())
	}
	
	// Test non-existent provider
	_, err = aggregator.GetProviderByName("NonExistentProvider")
	if err == nil {
		t.Error("Expected error when getting non-existent provider")
	}
}