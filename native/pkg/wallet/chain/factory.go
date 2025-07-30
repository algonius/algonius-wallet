// SPDX-License-Identifier: Apache-2.0
package chain

import (
	"fmt"
	"strings"
	"sync"

	"github.com/algonius/algonius-wallet/native/pkg/dex"
	"go.uber.org/zap"
)

// ChainFactory manages chain instances
type ChainFactory struct {
	chains        map[string]IChain
	dexAggregator dex.IDEXAggregator
	logger        *zap.Logger
	mu            sync.RWMutex
}

// NewChainFactory creates a new chain factory (legacy mode)
func NewChainFactory() *ChainFactory {
	factory := &ChainFactory{
		chains: make(map[string]IChain),
	}

	// Register available chains (legacy mode without DEX aggregator)
	factory.RegisterChain("ETH", NewETHChainLegacy())
	factory.RegisterChain("ETHEREUM", NewETHChainLegacy())
	factory.RegisterChain("BSC", NewBSCChainLegacy())
	factory.RegisterChain("BINANCE", NewBSCChainLegacy())
	factory.RegisterChain("SOL", NewSolanaChainLegacy())
	factory.RegisterChain("SOLANA", NewSolanaChainLegacy())

	return factory
}

// NewChainFactoryWithDEX creates a new chain factory with DEX aggregator support
func NewChainFactoryWithDEX(dexAggregator dex.IDEXAggregator, logger *zap.Logger) *ChainFactory {
	factory := &ChainFactory{
		chains:        make(map[string]IChain),
		dexAggregator: dexAggregator,
		logger:        logger,
	}

	// Register chains with DEX aggregator support
	factory.RegisterChain("ETH", NewETHChain(dexAggregator, logger))
	factory.RegisterChain("ETHEREUM", NewETHChain(dexAggregator, logger))
	factory.RegisterChain("BSC", NewBSCChain(dexAggregator, logger))
	factory.RegisterChain("BINANCE", NewBSCChain(dexAggregator, logger))
	// Handle potential error from NewSolanaChain
	if solanaChain, err := NewSolanaChain(dexAggregator, logger); err == nil {
		factory.RegisterChain("SOL", solanaChain)
		factory.RegisterChain("SOLANA", solanaChain)
	} else {
		// Fallback to legacy Solana chain if enhanced version fails
		logger.Warn("Failed to create enhanced Solana chain, using legacy version", zap.Error(err))
		legacyChain := NewSolanaChainLegacy()
		factory.RegisterChain("SOL", legacyChain)
		factory.RegisterChain("SOLANA", legacyChain)
	}

	return factory
}

// RegisterChain registers a new chain implementation
func (cf *ChainFactory) RegisterChain(name string, chain IChain) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	cf.chains[strings.ToUpper(name)] = chain
}

// GetChain returns a chain implementation by name
func (cf *ChainFactory) GetChain(name string) (IChain, error) {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	
	chain, exists := cf.chains[strings.ToUpper(name)]
	if !exists {
		return nil, fmt.Errorf("unsupported chain: %s", name)
	}
	return chain, nil
}

// GetSupportedChains returns a list of supported chain names
func (cf *ChainFactory) GetSupportedChains() []string {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	
	chains := make([]string, 0, len(cf.chains))
	seen := make(map[string]bool)

	for _, chain := range cf.chains {
		chainName := chain.GetChainName()
		if !seen[chainName] {
			chains = append(chains, chainName)
			seen[chainName] = true
		}
	}

	return chains
}

// SetDEXAggregator updates the DEX aggregator for the factory and re-registers chains
func (cf *ChainFactory) SetDEXAggregator(dexAggregator dex.IDEXAggregator, logger *zap.Logger) {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	cf.dexAggregator = dexAggregator
	cf.logger = logger

	// Re-register chains with DEX support
	cf.chains["ETH"] = NewETHChain(dexAggregator, logger)
	cf.chains["ETHEREUM"] = NewETHChain(dexAggregator, logger)
	cf.chains["BSC"] = NewBSCChain(dexAggregator, logger)
	cf.chains["BINANCE"] = NewBSCChain(dexAggregator, logger)
	// Handle potential error from NewSolanaChain
	if solanaChain, err := NewSolanaChain(dexAggregator, logger); err == nil {
		cf.chains["SOL"] = solanaChain
		cf.chains["SOLANA"] = solanaChain
	} else {
		// Fallback to legacy Solana chain if enhanced version fails
		if logger != nil {
			logger.Warn("Failed to create enhanced Solana chain, using legacy version", zap.Error(err))
		}
		legacyChain := NewSolanaChainLegacy()
		cf.chains["SOL"] = legacyChain
		cf.chains["SOLANA"] = legacyChain
	}

	if logger != nil {
		logger.Info("Chain factory updated with DEX aggregator support")
	}
}

// GetDEXAggregator returns the current DEX aggregator
func (cf *ChainFactory) GetDEXAggregator() dex.IDEXAggregator {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	return cf.dexAggregator
}

// HasDEXSupport returns true if the factory has DEX aggregator support
func (cf *ChainFactory) HasDEXSupport() bool {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	return cf.dexAggregator != nil
}