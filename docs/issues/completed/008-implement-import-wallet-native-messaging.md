---
title: 'Implement import_wallet Native Messaging RPC Method'
labels: ['enhancement', 'native-messaging', 'browser-extension', 'high-priority', 'security', 'completed']
assignees: []
---

## Summary

Implement the `import_wallet` RPC method for Native Messaging to enable browser extensions to securely import wallets using mnemonic phrases.

## Background

Browser extensions need the ability to import existing wallets using mnemonic phrases. This is a security-sensitive operation that should only be available through Native Messaging, not exposed to AI Agents via MCP.

## Requirements

### Functional Requirements

- [x] Implement secure mnemonic phrase validation
- [x] Support BIP39 standard mnemonic phrases (12/24 words)
- [x] Validate password strength requirements
- [x] Support multiple blockchain networks (ETH, BSC)
- [x] Implement secure private key storage
- [x] Support custom derivation paths
- [x] Prevent duplicate wallet imports

### Technical Requirements

- [x] Create `native/pkg/messaging/handlers/import_wallet_handler.go`
- [x] Implement JSON-RPC method following Native Messaging protocol
- [x] Integrate with existing wallet manager
- [x] Add comprehensive input validation
- [x] Implement proper error handling with specific error codes
- [x] Use secure encryption for private key storage

### Security Requirements

- [x] Validate mnemonic phrases against BIP39 standard
- [x] Enforce strong password requirements
- [x] Encrypt private keys before storage
- [x] Prevent exposure of sensitive data in logs
- [x] Implement rate limiting to prevent brute-force attacks
- [x] Validate derivation paths to prevent path traversal

## Acceptance Criteria

- [x] Method successfully imports wallets using valid mnemonic phrases
- [x] Proper error handling for invalid mnemonics, weak passwords, etc.
- [x] Private keys are securely encrypted before storage
- [x] Response includes wallet address and public key
- [x] Unit tests pass with various input scenarios
- [x] Security validations prevent unauthorized access
- [x] Integration with browser extension works correctly

## Implementation Details

### RPC Method Schema

```json
{
  "method": "import_wallet",
  "description": "Import an existing wallet using a mnemonic phrase via Native Messaging",
  "params": {
    "type": "object",
    "properties": {
      "mnemonic": {
        "type": "string",
        "description": "BIP39 mnemonic phrase (12 or 24 words)"
      },
      "password": {
        "type": "string",
        "description": "Password for encrypting private keys"
      },
      "chain": {
        "type": "string",
        "description": "Target blockchain network (ethereum, bsc)"
      },
      "derivation_path": {
        "type": "string",
        "description": "Custom derivation path (optional, defaults to standard path)"
      }
    },
    "required": ["mnemonic", "password", "chain"]
  },
  "result": {
    "type": "object",
    "properties": {
      "address": { "type": "string" },
      "public_key": { "type": "string" },
      "imported_at": { "type": "integer" }
    },
    "required": ["address", "public_key", "imported_at"]
  }
}
```

### Files Created/Modified

- `native/pkg/messaging/handlers/import_wallet_handler.go` (new)
- `native/pkg/messaging/handlers/import_wallet_handler_test.go` (new)
- `native/pkg/wallet/manager.go` (extended with ImportWallet method)
- `native/pkg/wallet/validation.go` (extended with mnemonic and password validation)
- `native/pkg/security/crypto.go` (used for encryption)
- `native/cmd/main.go` (registered RPC method)

### Error Codes

- `-32001`: Invalid mnemonic phrase
- `-32002`: Weak password
- `-32003`: Unsupported chain
- `-32004`: Wallet already exists
- `-32005`: Storage encryption failed
- `-32602`: Invalid parameters

### Response Format

The method returns a JSON-RPC response with:

1. Wallet address
2. Public key
3. Import timestamp

### Key Features Implemented

1. **Secure Import**: Validates mnemonic phrases against BIP39 standard
2. **Multi-chain Support**: Works with Ethereum and BSC
3. **Strong Security**: Enforces password requirements and encrypts private keys
4. **Custom Derivation**: Supports custom derivation paths
5. **Error Handling**: Provides specific error codes for different failure scenarios
6. **Prevent Duplicates**: Prevents importing the same wallet multiple times

## Dependencies

- Requires secure storage implementation
- Depends on BIP39 mnemonic validation library
- Related to wallet manager and chain interfaces

## Testing Requirements

- [x] Unit tests for mnemonic validation
- [x] Integration tests with various mnemonic formats
- [x] Security tests for key encryption/decryption
- [x] Error case testing (invalid mnemonics, weak passwords)
- [x] Performance tests for large wallet imports

## References

- Technical Spec: `docs/teck_spec.md`
- Native Messaging: `native/pkg/messaging/native.go`
- BIP39 Standard: https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki
