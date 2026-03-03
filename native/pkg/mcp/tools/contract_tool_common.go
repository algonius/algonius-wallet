package tools

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mr-tron/base58"
)

func isValidAddressForChain(chainName, address string) bool {
	switch chainName {
	case "ethereum", "bsc":
		return common.IsHexAddress(address)
	case "solana":
		if address == "" {
			return false
		}
		decoded, err := base58.Decode(address)
		return err == nil && len(decoded) > 0
	default:
		return false
	}
}

func validateBytecode(chainName, bytecode string) error {
	bytecode = strings.TrimSpace(bytecode)
	if bytecode == "" {
		return fmt.Errorf("bytecode cannot be empty")
	}
	if chainName == "ethereum" || chainName == "bsc" {
		if !strings.HasPrefix(bytecode, "0x") || len(bytecode) <= 2 {
			return fmt.Errorf("EVM bytecode must be a non-empty 0x-prefixed hex string")
		}
	}
	return nil
}

func deterministicTxHash(chainName, seed string) string {
	raw := crypto.Keccak256([]byte(seed))
	if chainName == "solana" {
		return base58.Encode(raw)
	}
	return common.BytesToHash(raw).Hex()
}

func deterministicContractAddress(chainName, seed string) string {
	raw := crypto.Keccak256([]byte(seed + "|contract"))
	if chainName == "solana" {
		return base58.Encode(raw[:32])
	}
	return common.BytesToAddress(raw[12:]).Hex()
}

func deterministicCallResult(chainName, method, seed string) string {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "balanceof", "getbalance":
		return "0"
	case "symbol":
		return "ALG"
	case "decimals":
		return "18"
	}

	raw := crypto.Keccak256([]byte(seed))
	if chainName == "solana" {
		return base58.Encode(raw[:32])
	}
	return "0x" + hex.EncodeToString(raw[:32])
}
