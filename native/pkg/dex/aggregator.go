// SPDX-License-Identifier: Apache-2.0
package dex

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"go.uber.org/zap"
)

// DEXProviderConfig contains provider-specific configuration
type DEXProviderConfig struct {
	Name        string                 `json:"name"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`    // Higher priority = preferred
	Config      map[string]interface{} `json:"config"`      // Provider-specific settings
	SupportedChains []string           `json:"supported_chains"`
}

// DEXAggregator implements the IDEXAggregator interface
type DEXAggregator struct {
	providers map[string]IDEXProvider
	configs   map[string]*DEXProviderConfig
	logger    *zap.Logger
	mu        sync.RWMutex
}

// NewDEXAggregator creates a new DEX aggregator instance
func NewDEXAggregator(logger *zap.Logger) *DEXAggregator {
	return &DEXAggregator{
		providers: make(map[string]IDEXProvider),
		configs:   make(map[string]*DEXProviderConfig),
		logger:    logger,
	}
}

// RegisterProvider registers a new DEX provider
func (d *DEXAggregator) RegisterProvider(provider IDEXProvider) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	name := provider.GetName()
	if _, exists := d.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	d.providers[name] = provider
	d.logger.Info("Registered DEX provider", zap.String("provider", name))
	return nil
}

// RegisterProviderWithConfig registers a provider with configuration
func (d *DEXAggregator) RegisterProviderWithConfig(provider IDEXProvider, config *DEXProviderConfig) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	name := provider.GetName()
	if _, exists := d.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	d.providers[name] = provider
	d.configs[name] = config
	d.logger.Info("Registered DEX provider with config", 
		zap.String("provider", name),
		zap.Bool("enabled", config.Enabled),
		zap.Int("priority", config.Priority))
	return nil
}

// GetBestQuote gets the best quote from all available providers
func (d *DEXAggregator) GetBestQuote(ctx context.Context, params SwapParams) (*SwapQuote, error) {
	d.mu.RLock()
	supportedProviders := d.getSupportedProviders(params.ChainID)
	d.mu.RUnlock()

	if len(supportedProviders) == 0 {
		return nil, fmt.Errorf("no providers support chain %s", params.ChainID)
	}

	// Channel to collect quotes from all providers
	type quoteResult struct {
		quote *SwapQuote
		err   error
		provider string
	}
	
	quoteChan := make(chan quoteResult, len(supportedProviders))
	
	// Request quotes from all supported providers concurrently
	for _, providerName := range supportedProviders {
		go func(name string) {
			provider := d.providers[name]
			quote, err := provider.GetQuote(ctx, params)
			if quote != nil {
				quote.Provider = name
			}
			quoteChan <- quoteResult{quote: quote, err: err, provider: name}
		}(providerName)
	}

	// Collect quotes
	var bestQuote *SwapQuote
	var quotes []*SwapQuote
	var errors []error

	for i := 0; i < len(supportedProviders); i++ {
		result := <-quoteChan
		if result.err != nil {
			d.logger.Warn("Provider quote failed", 
				zap.String("provider", result.provider),
				zap.Error(result.err))
			errors = append(errors, result.err)
			continue
		}
		
		if result.quote != nil {
			quotes = append(quotes, result.quote)
			d.logger.Debug("Received quote", 
				zap.String("provider", result.provider),
				zap.String("toAmount", result.quote.ToAmount))
		}
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("no valid quotes received, errors: %v", errors)
	}

	// Find the best quote (highest output amount)
	bestQuote = d.selectBestQuote(quotes)
	
	d.logger.Info("Selected best quote", 
		zap.String("provider", bestQuote.Provider),
		zap.String("fromAmount", bestQuote.FromAmount),
		zap.String("toAmount", bestQuote.ToAmount),
		zap.Float64("slippage", bestQuote.Slippage))

	return bestQuote, nil
}

// selectBestQuote selects the best quote based on output amount and provider priority
func (d *DEXAggregator) selectBestQuote(quotes []*SwapQuote) *SwapQuote {
	if len(quotes) == 1 {
		return quotes[0]
	}

	// Sort quotes by output amount (descending) and provider priority
	sort.Slice(quotes, func(i, j int) bool {
		// Parse output amounts for comparison
		amountI, errI := strconv.ParseFloat(quotes[i].ToAmount, 64)
		amountJ, errJ := strconv.ParseFloat(quotes[j].ToAmount, 64)
		
		if errI != nil || errJ != nil {
			// Fallback to provider priority if amount parsing fails
			return d.getProviderPriority(quotes[i].Provider) > d.getProviderPriority(quotes[j].Provider)
		}
		
		// If amounts are very close (within 0.1%), use provider priority
		if abs(amountI-amountJ)/amountI < 0.001 {
			return d.getProviderPriority(quotes[i].Provider) > d.getProviderPriority(quotes[j].Provider)
		}
		
		return amountI > amountJ
	})

	return quotes[0]
}

// getProviderPriority returns the priority of a provider
func (d *DEXAggregator) getProviderPriority(providerName string) int {
	if config, exists := d.configs[providerName]; exists {
		return config.Priority
	}
	return 0 // Default priority
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// ExecuteSwapWithProvider executes swap using a specific provider
func (d *DEXAggregator) ExecuteSwapWithProvider(ctx context.Context, providerName string, params SwapParams) (*SwapResult, error) {
	d.mu.RLock()
	provider, exists := d.providers[providerName]
	d.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	if !provider.IsSupported(params.ChainID) {
		return nil, fmt.Errorf("provider %s does not support chain %s", providerName, params.ChainID)
	}

	d.logger.Info("Executing swap with provider", 
		zap.String("provider", providerName),
		zap.String("fromToken", params.FromToken),
		zap.String("toToken", params.ToToken),
		zap.String("amount", params.Amount))

	result, err := provider.ExecuteSwap(ctx, params)
	if err != nil {
		d.logger.Error("Swap execution failed", 
			zap.String("provider", providerName),
			zap.Error(err))
		return nil, err
	}

	d.logger.Info("Swap executed successfully", 
		zap.String("provider", providerName),
		zap.String("txHash", result.TxHash))

	return result, nil
}

// GetSupportedProviders returns list of providers supporting the chain
func (d *DEXAggregator) GetSupportedProviders(chainID string) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.getSupportedProviders(chainID)
}

// getSupportedProviders internal method to get supported providers (assumes lock is held)
func (d *DEXAggregator) getSupportedProviders(chainID string) []string {
	var supported []string
	
	for name, provider := range d.providers {
		// Check if provider is enabled in config
		if config, exists := d.configs[name]; exists && !config.Enabled {
			continue
		}
		
		if provider.IsSupported(chainID) {
			supported = append(supported, name)
		}
	}
	
	// Sort by priority (descending)
	sort.Slice(supported, func(i, j int) bool {
		priorityI := d.getProviderPriority(supported[i])
		priorityJ := d.getProviderPriority(supported[j])
		return priorityI > priorityJ
	})
	
	return supported
}

// GetProviderByName returns a specific provider by name
func (d *DEXAggregator) GetProviderByName(name string) (IDEXProvider, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	provider, exists := d.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}