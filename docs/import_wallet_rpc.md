# Import Wallet Native Messaging RPC Method

## Overview

The `import_wallet` RPC method allows browser extensions to securely import existing wallets using BIP39 mnemonic phrases. This method is **only available through Native Messaging** and is never exposed to AI Agents via MCP tools, maintaining strict security boundaries.

## Method Specification

### Request Schema

```json
{
  "id": "string",
  "method": "import_wallet",
  "params": {
    "mnemonic": "string (required)",
    "password": "string (required)",
    "chain": "string (required)",
    "derivation_path": "string (optional)"
  }
}
```

### Response Schema

**Success Response:**

```json
{
  "id": "string",
  "result": {
    "address": "string",
    "public_key": "string",
    "imported_at": "number (timestamp)"
  }
}
```

**Error Response:**

```json
{
  "id": "string",
  "error": {
    "code": "number",
    "message": "string"
  }
}
```

## Parameters

| Parameter         | Type   | Required | Description                                                    |
| ----------------- | ------ | -------- | -------------------------------------------------------------- |
| `mnemonic`        | string | Yes      | BIP39 mnemonic phrase (12, 15, 18, 21, or 24 words)            |
| `password`        | string | Yes      | Password for encrypting private key storage (min 8 chars)      |
| `chain`           | string | Yes      | Target blockchain: `"ethereum"`, `"eth"`, `"bsc"`, `"binance"` |
| `derivation_path` | string | No       | HD wallet derivation path (default: `"m/44'/60'/0'/0/0"`)      |

## Supported Chains

- **Ethereum** (`"ethereum"`, `"eth"`)
- **Binance Smart Chain** (`"bsc"`, `"binance"`)

Both chains use Ethereum-compatible addressing and are interoperable.

## Error Codes

| Code   | Constant                     | Description                                 |
| ------ | ---------------------------- | ------------------------------------------- |
| -32001 | `ErrInvalidMnemonic`         | Invalid or malformed mnemonic phrase        |
| -32002 | `ErrWeakPassword`            | Password doesn't meet security requirements |
| -32003 | `ErrUnsupportedChain`        | Chain not supported for import              |
| -32004 | `ErrWalletAlreadyExists`     | Wallet already exists in storage            |
| -32005 | `ErrStorageEncryptionFailed` | Failed to encrypt data for storage          |
| -32602 | -                            | Invalid request parameters                  |

## Security Features

### Mnemonic Validation

- Validates BIP39 mnemonic phrase format and checksum
- Supports standard word counts: 12, 15, 18, 21, 24 words
- Case-insensitive with whitespace normalization

### Password Security

- Minimum 8 character requirement
- Used for AES-256-GCM encryption of private keys
- Combined with PBKDF2 (100,000 iterations) for key derivation

### Private Key Protection

- Private keys encrypted before any storage
- Only public information (address, public key) returned
- Private keys never logged or exposed in responses

### Input Sanitization

- All inputs validated and sanitized
- Chain names normalized to standard format
- Derivation paths validated for security

## Usage Examples

### Basic Import (Ethereum)

```json
{
  "id": "import-001",
  "method": "import_wallet",
  "params": {
    "mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
    "password": "my_secure_password_123",
    "chain": "ethereum"
  }
}
```

**Response:**

```json
{
  "id": "import-001",
  "result": {
    "address": "0x9858EfFD232B4033E47d90003D41EC34EcaEda94",
    "public_key": "0x04c9c6c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7c7",
    "imported_at": 1640995200
  }
}
```

### Import with Custom Derivation Path

```json
{
  "id": "import-002",
  "method": "import_wallet",
  "params": {
    "mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
    "password": "secure_password_456",
    "chain": "bsc",
    "derivation_path": "m/44'/60'/0'/0/1"
  }
}
```

### Error Example - Invalid Mnemonic

```json
{
  "id": "import-003",
  "method": "import_wallet",
  "params": {
    "mnemonic": "invalid mnemonic phrase",
    "password": "password123",
    "chain": "ethereum"
  }
}
```

**Error Response:**

```json
{
  "id": "import-003",
  "error": {
    "code": -32001,
    "message": "invalid mnemonic: invalid mnemonic phrase"
  }
}
```

## Implementation Details

### Architecture

- Native Messaging mode automatically detected via stdin pipe
- Shared wallet manager between MCP and Native Messaging modes
- Import functionality completely isolated from MCP tools

### Encryption

- AES-256-GCM with PBKDF2 key derivation
- Unique salt per encryption operation
- 100,000 PBKDF2 iterations for key strengthening

### Chain Support

- Utilizes existing chain factory architecture
- Ethereum and BSC chains supported
- Extensible for additional EVM-compatible chains

## Testing

### Unit Tests

- Mnemonic validation with various formats
- Password strength requirements
- Chain validation and normalization
- Encryption/decryption functionality
- RPC handler error scenarios

### Integration Tests

- End-to-end wallet import flow
- Cross-chain compatibility
- Security boundary validation
- Error handling verification

### Test Coverage

```bash
# Run all tests
make test

# Run specific import tests
go test -v ./pkg/wallet/validation_test.go
go test -v ./pkg/security/crypto_test.go
go test -v ./pkg/messaging/import_wallet_handler_test.go
```

## Security Considerations

### Browser Extension Only

- Only accessible via Native Messaging protocol
- Never exposed to AI Agents or MCP tools
- Maintains strict security boundaries per system design

### Private Key Handling

- Private keys never stored in plain text
- Encryption occurs before any storage operations
- Keys derived but never logged or transmitted

### Rate Limiting (Future)

- Consider implementing rate limiting for import attempts
- Prevent brute force attacks on weak passwords
- Add user confirmation mechanisms for production

## Future Enhancements

1. **HD Wallet Derivation**: Implement proper BIP32 hierarchical deterministic wallet derivation
2. **Hardware Wallet Support**: Integration with hardware wallets for enhanced security
3. **Multi-Account Import**: Support importing multiple accounts from single mnemonic
4. **Backup Verification**: Verify imported mnemonic matches existing backup
5. **Advanced Validation**: Additional mnemonic phrase entropy and security checks
