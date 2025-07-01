// SPDX-License-Identifier: Apache-2.0
package wallet

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/wallet/chain"
)

// TransactionCache provides caching for transaction confirmations
type TransactionCache struct {
	cache map[string]*CachedConfirmation
	mutex sync.RWMutex
	ttl   time.Duration
}

// CachedConfirmation represents a cached transaction confirmation with expiry
type CachedConfirmation struct {
	Confirmation *chain.TransactionConfirmation
	ExpiresAt    time.Time
}

// NewTransactionCache creates a new transaction cache with specified TTL
func NewTransactionCache(ttl time.Duration) *TransactionCache {
	return &TransactionCache{
		cache: make(map[string]*CachedConfirmation),
		ttl:   ttl,
	}
}

// Get retrieves a cached confirmation if it exists and hasn't expired
func (tc *TransactionCache) Get(key string) (*chain.TransactionConfirmation, bool) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	cached, exists := tc.cache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(cached.ExpiresAt) {
		// Cleanup expired entry
		delete(tc.cache, key)
		return nil, false
	}

	return cached.Confirmation, true
}

// Set stores a confirmation in the cache with TTL
func (tc *TransactionCache) Set(key string, confirmation *chain.TransactionConfirmation) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	tc.cache[key] = &CachedConfirmation{
		Confirmation: confirmation,
		ExpiresAt:    time.Now().Add(tc.ttl),
	}
}

// CleanupExpired removes expired entries from the cache
func (tc *TransactionCache) CleanupExpired() {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	now := time.Now()
	for key, cached := range tc.cache {
		if now.After(cached.ExpiresAt) {
			delete(tc.cache, key)
		}
	}
}

// Global cache instance with 30-second TTL for transaction confirmations
var transactionCache = NewTransactionCache(30 * time.Second)

// ConfirmTransaction checks transaction confirmation status with caching
func ConfirmTransaction(ctx context.Context, chainName, txHash string, requiredConfirmations uint64, factory *chain.ChainFactory) (*chain.TransactionConfirmation, error) {
	if chainName == "" {
		return nil, errors.New("chain name is required")
	}
	if txHash == "" {
		return nil, errors.New("transaction hash is required")
	}

	// Create cache key
	cacheKey := strings.ToLower(chainName) + ":" + strings.ToLower(txHash)

	// Check cache first
	if cached, found := transactionCache.Get(cacheKey); found {
		// For pending transactions, don't use cache to get real-time updates
		if cached.Status != "pending" {
			return cached, nil
		}
	}

	// Get chain implementation
	chainImpl, err := factory.GetChain(chainName)
	if err != nil {
		return nil, err
	}

	// Query blockchain for confirmation status
	confirmation, err := chainImpl.ConfirmTransaction(ctx, txHash, requiredConfirmations)
	if err != nil {
		return nil, err
	}

	// Cache the result (with shorter TTL for pending transactions)
	if confirmation.Status == "pending" {
		// Don't cache pending transactions as aggressively
		shortCache := NewTransactionCache(5 * time.Second)
		shortCache.Set(cacheKey, confirmation)
	} else {
		// Cache confirmed/failed transactions longer
		transactionCache.Set(cacheKey, confirmation)
	}

	return confirmation, nil
}

// SendTransaction sends a transaction from one address to another.
// Currently only supports ETH, always returns a mock tx_hash.
func SendTransaction(from, to, amount, token, chain string) (string, error) {
	if from == "" || to == "" || amount == "" {
		return "", errors.New("from, to, and amount are required")
	}
	if strings.ToUpper(token) != "ETH" && token != "" {
		return "", errors.New("unsupported token: only ETH is supported")
	}
	if strings.ToUpper(chain) != "ETH" && chain != "" {
		return "", errors.New("unsupported chain: only ETH is supported")
	}
	// TODO: Integrate with blockchain node or provider
	return "0xMOCKTXHASH", nil
}
