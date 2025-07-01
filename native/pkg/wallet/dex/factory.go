// SPDX-License-Identifier: Apache-2.0
package dex

import (
	"fmt"
	"strings"
)

// DEXFactory implements IDEXFactory for creating DEX instances
type DEXFactory struct{}

// NewDEXFactory creates a new DEX factory
func NewDEXFactory() *DEXFactory {
	return &DEXFactory{}
}

// CreateDEX creates a DEX instance for the specified protocol and chain
func (f *DEXFactory) CreateDEX(protocol string, chain string) (IDEX, error) {
	protocol = strings.ToLower(protocol)
	chain = strings.ToLower(chain)
	
	switch protocol {
	case "uniswap", "uniswap_v2":
		if chain != "ethereum" && chain != "eth" {
			return nil, fmt.Errorf("uniswap is only supported on Ethereum, got chain: %s", chain)
		}
		return NewUniswapV2(chain), nil
		
	case "pancakeswap":
		if chain != "bsc" && chain != "binance" {
			return nil, fmt.Errorf("pancakeswap is only supported on BSC, got chain: %s", chain)
		}
		// TODO: Implement PancakeSwap in future phase
		return nil, fmt.Errorf("pancakeswap support not implemented yet")
		
	default:
		return nil, fmt.Errorf("unsupported DEX protocol: %s", protocol)
	}
}

// GetSupportedProtocols returns supported DEX protocols
func (f *DEXFactory) GetSupportedProtocols() []string {
	return []string{
		"uniswap",
		"uniswap_v2",
		// "pancakeswap", // TODO: Add in future phase
	}
}

// GetSupportedChains returns supported blockchain networks for a protocol
func (f *DEXFactory) GetSupportedChains(protocol string) []string {
	protocol = strings.ToLower(protocol)
	
	switch protocol {
	case "uniswap", "uniswap_v2":
		return []string{"ethereum"}
		
	case "pancakeswap":
		return []string{"bsc"}
		
	default:
		return []string{}
	}
}