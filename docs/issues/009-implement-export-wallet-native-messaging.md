---
title: 'Implement export_wallet Native Messaging RPC Method'
labels: ['enhancement', 'native-messaging', 'browser-extension', 'high-priority', 'security']
assignees: []
---

## Summary

Implement the `export_wallet` RPC method for Native Messaging to enable browser extensions to securely export wallet private keys or mnemonic phrases.

## Background

Browser extensions need the ability to export wallet private keys or mnemonic phrases for backup purposes. This is a highly security-sensitive operation that requires strong authentication and should only be available through Native Messaging.

## Requirements

### Functional Requirements

- [ ] Export private key for specific wallet address
- [ ] Export mnemonic phrase for HD wallets
- [ ] Support multiple export formats (private key, mnemonic)
- [ ] Require password authentication before export
- [ ] Return encrypted export data

### Technical Requirements

- [ ] Add RPC method handler in `native/pkg/messaging/native.go`
- [ ] Integrate with wallet manager for key retrieval
- [ ] Implement strong authentication mechanisms
- [ ] Add export format validation
- [ ] Support temporary export sessions

### Security Requirements

- [ ] Require master password verification
- [ ] Implement export rate limiting
- [ ] Add audit logging for export operations
- [ ] Encrypt export data with user-provided key
- [ ] Never log or cache exported private data
- [ ] Implement export session timeouts

## Acceptance Criteria

- [ ] RPC method authenticates user before export
- [ ] Successfully exports private keys and mnemonics
- [ ] Proper error handling for invalid passwords
- [ ] Export data is properly encrypted
- [ ] Audit trail records all export attempts
- [ ] Integration tests pass with encrypted exports

## Implementation Details

### RPC Method Schema

```json
{
  "method": "export_wallet",
  "params": {
    "address": "string (required)",
    "password": "string (required)",
    "export_type": "string (required, enum: [\"private_key\", \"mnemonic\"])",
    "encryption_key": "string (optional)"
  },
  "result": {
    "export_data": "string (encrypted)",
    "export_type": "string",
    "exported_at": "number (timestamp)",
    "expires_at": "number (timestamp)"
  },
  "error": {
    "code": "number",
    "message": "string"
  }
}
```

### Files to Modify

- `native/pkg/messaging/native.go` - Add RPC method registration
- `native/pkg/wallet/manager.go` - Add export functionality
- `native/pkg/wallet/crypto.go` - Add export encryption utilities
- `native/pkg/wallet/auth.go` - Add authentication mechanisms
- `native/pkg/audit/logger.go` - Add audit logging

### Error Codes

- `-32011`: Invalid password
- `-32012`: Wallet not found
- `-32013`: Export type not supported
- `-32014`: Too many export attempts
- `-32015`: Export encryption failed

## Dependencies

- Requires authentication system
- Depends on audit logging framework
- Related to secure storage and encryption utilities

## Testing Requirements

- [ ] Unit tests for password authentication
- [ ] Integration tests with various export types
- [ ] Security tests for export encryption
- [ ] Error case testing (wrong passwords, missing wallets)
- [ ] Rate limiting tests

## References

- Technical Spec: `docs/teck_spec.md`
- Native Messaging: `native/pkg/messaging/native.go`
- Related Issue: #008 (import_wallet)
