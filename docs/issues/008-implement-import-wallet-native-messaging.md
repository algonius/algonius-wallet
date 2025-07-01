---
title: 'Implement import_wallet Native Messaging RPC Method'
labels: ['enhancement', 'native-messaging', 'browser-extension', 'high-priority', 'security']
assignees: []
---

## Summary

Implement the `import_wallet` RPC method for Native Messaging to enable browser extensions to securely import wallets using mnemonic phrases.

## Background

Browser extensions need the ability to import existing wallets using mnemonic phrases. This is a security-sensitive operation that should only be available through Native Messaging, not exposed to AI Agents via MCP.

## Requirements

### Functional Requirements

- [ ] Implement secure mnemonic phrase validation
- [ ] Support BIP39 standard mnemonic phrases (12/24 words)
- [ ] Generate wallet addresses for supported chains (ETH, BSC)
- [ ] Encrypt and store private keys securely
- [ ] Return wallet address and public key only

### Technical Requirements

- [ ] Add RPC method handler in `native/pkg/messaging/native.go`
- [ ] Integrate with existing wallet manager
- [ ] Implement proper input validation and sanitization
- [ ] Add secure key storage mechanisms
- [ ] Support multiple derivation paths

### Security Requirements

- [ ] Validate mnemonic phrase format and checksum
- [ ] Encrypt private keys before storage
- [ ] Never log or expose private keys
- [ ] Implement rate limiting for import attempts
- [ ] Add user confirmation mechanisms

## Acceptance Criteria

- [ ] RPC method accepts valid mnemonic phrases
- [ ] Successfully generates and stores wallet for supported chains
- [ ] Proper error handling for invalid mnemonics
- [ ] Private keys are encrypted and securely stored
- [ ] Returns only public information (address, public key)
- [ ] Integration tests pass with real mnemonic phrases

## Implementation Details

### RPC Method Schema

```json
{
  "method": "import_wallet",
  "params": {
    "mnemonic": "string (required)",
    "password": "string (required)",
    "chain": "string (required, enum: [\"ethereum\", \"bsc\"])",
    "derivation_path": "string (optional, default: \"m/44'/60'/0'/0/0\")"
  },
  "result": {
    "address": "string",
    "public_key": "string",
    "imported_at": "number (timestamp)"
  },
  "error": {
    "code": "number",
    "message": "string"
  }
}
```

### Files to Modify

- `native/pkg/messaging/native.go` - Add RPC method registration
- `native/pkg/wallet/manager.go` - Add import functionality
- `native/pkg/wallet/crypto.go` - Add encryption/decryption utilities
- `native/pkg/wallet/validation.go` - Add mnemonic validation

### Error Codes

- `-32001`: Invalid mnemonic phrase
- `-32002`: Weak password
- `-32003`: Unsupported chain
- `-32004`: Wallet already exists
- `-32005`: Storage encryption failed

## Dependencies

- Requires secure storage implementation
- Depends on BIP39 mnemonic validation library
- Related to wallet manager and chain interfaces

## Testing Requirements

- [ ] Unit tests for mnemonic validation
- [ ] Integration tests with various mnemonic formats
- [ ] Security tests for key encryption/decryption
- [ ] Error case testing (invalid mnemonics, weak passwords)
- [ ] Performance tests for large wallet imports

## References

- Technical Spec: `docs/teck_spec.md`
- Native Messaging: `native/pkg/messaging/native.go`
- BIP39 Standard: https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki
