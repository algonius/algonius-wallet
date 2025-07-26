// SPDX-License-Identifier: Apache-2.0
package chain

import (
	"fmt"
	"strings"
)

// ChainFactory manages chain instances
type ChainFactory struct {
	chains map[string]IChain
}

// NewChainFactory creates a new chain factory
func NewChainFactory() *ChainFactory {
	factory := &ChainFactory{
		chains: make(map[string]IChain),
	}

	// Register available chains
	factory.RegisterChain("ETH", NewETHChain())
	factory.RegisterChain("ETHEREUM", NewETHChain())
	factory.RegisterChain("BSC", NewBSCChain())
	factory.RegisterChain("BINANCE", NewBSCChain())
	factory.RegisterChain("SOL", NewSolanaChain())
	factory.RegisterChain("SOLANA", NewSolanaChain())

	return factory
}

// RegisterChain registers a new chain implementation
func (cf *ChainFactory) RegisterChain(name string, chain IChain) {
	cf.chains[strings.ToUpper(name)] = chain
}

// GetChain returns a chain implementation by name
func (cf *ChainFactory) GetChain(name string) (IChain, error) {
	chain, exists := cf.chains[strings.ToUpper(name)]
	if !exists {
		return nil, fmt.Errorf("unsupported chain: %s", name)
	}
	return chain, nil
}

// GetSupportedChains returns a list of supported chain names
func (cf *ChainFactory) GetSupportedChains() []string {
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
