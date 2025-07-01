// SPDX-License-Identifier: Apache-2.0
package wallet

import (
	"crypto/ecdsa"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// Wallet represents a blockchain wallet with address and public key.
type Wallet struct {
	Chain      string
	Address    string
	PublicKey  string
	PrivateKey *ecdsa.PrivateKey // Not exported outside package
}

// CreateWallet generates a new wallet for the specified chain.
// Currently only supports "ETH" (Ethereum).
func CreateWallet(chain string) (*Wallet, error) {
	if strings.ToUpper(chain) != "ETH" {
		return nil, errors.New("unsupported chain: only ETH is supported")
	}
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	publicKey := privateKey.PublicKey
	address := crypto.PubkeyToAddress(publicKey).Hex()
	pubBytes := crypto.FromECDSAPub(&publicKey)
	return &Wallet{
		Chain:      "ETH",
		Address:    address,
		PublicKey:  hexutil.Encode(pubBytes),
		PrivateKey: privateKey,
	}, nil
}
