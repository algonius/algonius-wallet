// SPDX-License-Identifier: Apache-2.0
package wallet

import "time"

// WalletStatus represents the current status of a wallet.
type WalletStatus struct {
	Address   string            `json:"address"`
	PublicKey string            `json:"public_key"`
	Ready     bool              `json:"ready"`
	Chains    map[string]bool   `json:"chains,omitempty"`
	LastUsed  int64             `json:"last_used,omitempty"`
}

// NewWalletStatus creates a new WalletStatus with default values.
func NewWalletStatus(address, publicKey string) *WalletStatus {
	return &WalletStatus{
		Address:   address,
		PublicKey: publicKey,
		Ready:     true, // Default to ready for new wallets
		Chains:    make(map[string]bool),
		LastUsed:  time.Now().Unix(),
	}
}
