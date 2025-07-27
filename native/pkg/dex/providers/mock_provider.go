package providers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/dex"
	"go.uber.org/zap"
)

// MockProvider implements IDEXProvider for testing purposes
type MockProvider struct {
	name           string
	supportedChains map[string]bool
	logger         *zap.Logger
	
	// Mock configuration
	shouldFailQuote   bool
	shouldFailSwap    bool
	shouldFailBalance bool
	customQuoteAmount string
}

// MockConfig holds configuration for mock provider
type MockConfig struct {
	Name              string
	SupportedChains   []string
	ShouldFailQuote   bool
	ShouldFailSwap    bool
	ShouldFailBalance bool
	CustomQuoteAmount string
}

// NewMockProvider creates a new mock DEX provider for testing
func NewMockProvider(config MockConfig, logger *zap.Logger) *MockProvider {
	if config.Name == "" {
		config.Name = "MockDEX"
	}
	
	supportedChains := make(map[string]bool)
	if len(config.SupportedChains) == 0 {
		// Default supported chains
		supportedChains["1"] = true   // Ethereum
		supportedChains["56"] = true  // BSC
		supportedChains["501"] = true // Solana
	} else {
		for _, chainID := range config.SupportedChains {
			supportedChains[chainID] = true
		}
	}

	return &MockProvider{
		name:              config.Name,
		supportedChains:   supportedChains,
		logger:            logger,
		shouldFailQuote:   config.ShouldFailQuote,
		shouldFailSwap:    config.ShouldFailSwap,
		shouldFailBalance: config.ShouldFailBalance,
		customQuoteAmount: config.CustomQuoteAmount,
	}
}

// GetName returns the provider name
func (m *MockProvider) GetName() string {
	return m.name
}

// IsSupported checks if the chain is supported
func (m *MockProvider) IsSupported(chainID string) bool {
	return m.supportedChains[chainID]
}

// GetQuote returns a mock quote
func (m *MockProvider) GetQuote(ctx context.Context, params dex.SwapParams) (*dex.SwapQuote, error) {
	if m.shouldFailQuote {
		return nil, fmt.Errorf("mock quote failure")
	}

	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid swap parameters: %w", err)
	}

	// Generate mock quote amount
	quoteAmount := m.customQuoteAmount
	if quoteAmount == "" {
		// Parse input amount and simulate 1:1000 conversion rate
		if amount, err := strconv.ParseFloat(params.Amount, 64); err == nil {
			quoteAmount = fmt.Sprintf("%.6f", amount*1000)
		} else {
			quoteAmount = "1000.0"
		}
	}

	return &dex.SwapQuote{
		Provider:     m.name,
		FromToken:    params.FromToken,
		ToToken:      params.ToToken,
		FromAmount:   params.Amount,
		ToAmount:     quoteAmount,
		EstimatedGas: 200000,
		GasPrice:     "20000000000", // 20 gwei
		Slippage:     params.Slippage,
		ValidUntil:   time.Now().Add(30 * time.Second).Unix(),
		RawData:      fmt.Sprintf(`{"mock_quote": true, "provider": "%s"}`, m.name),
	}, nil
}

// ExecuteSwap executes a mock swap
func (m *MockProvider) ExecuteSwap(ctx context.Context, params dex.SwapParams) (*dex.SwapResult, error) {
	if m.shouldFailSwap {
		return nil, fmt.Errorf("mock swap failure")
	}

	// Get quote first
	quote, err := m.GetQuote(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	// Generate mock transaction hash
	mockTxHash := fmt.Sprintf("0x%x%x", time.Now().Unix(), time.Now().Nanosecond())

	return &dex.SwapResult{
		TxHash:     mockTxHash,
		Provider:   m.name,
		FromToken:  params.FromToken,
		ToToken:    params.ToToken,
		FromAmount: params.Amount,
		ToAmount:   quote.ToAmount,
		Status:     "confirmed", // Mock as immediately confirmed
		Timestamp:  time.Now().Unix(),
	}, nil
}

// GetBalance returns mock balance information
func (m *MockProvider) GetBalance(ctx context.Context, address string, tokenAddress string, chainID string) (*dex.BalanceInfo, error) {
	if m.shouldFailBalance {
		return nil, fmt.Errorf("mock balance failure")
	}

	return &dex.BalanceInfo{
		TokenAddress: tokenAddress,
		TokenSymbol:  "MOCK",
		Balance:      "1000000000000000000", // 1.0 token with 18 decimals
		Decimals:     18,
		USDValue:     "1000.00",
	}, nil
}

// EstimateGas returns mock gas estimation
func (m *MockProvider) EstimateGas(ctx context.Context, params dex.SwapParams) (gasLimit uint64, gasPrice string, err error) {
	if m.shouldFailQuote { // Reuse quote failure flag for gas estimation
		return 0, "", fmt.Errorf("mock gas estimation failure")
	}

	return 200000, "20000000000", nil // 200k gas limit, 20 gwei gas price
}