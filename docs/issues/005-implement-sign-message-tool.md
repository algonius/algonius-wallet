---
title: 'Implement sign_message MCP Tool'
labels: ['enhancement', 'MCP', 'native-host', 'security', 'medium-priority']
assignees: []
---

## Summary

Implement the `sign_message` MCP tool to enable AI Agents to cryptographically sign messages for authentication and authorization purposes.

## Background

Message signing is crucial for decentralized application authentication, proving wallet ownership, and secure communication protocols. This tool enables AI agents to interact with dApps that require message signatures.

## Requirements

### Functional Requirements

- [ ] Sign arbitrary text messages using wallet private key
- [ ] Support EIP-191 and EIP-712 signing standards
- [ ] Implement typed data signing for structured messages
- [ ] Support multiple signature formats (raw, hex, base64)
- [ ] Validate message format and content before signing
- [ ] Support batch message signing

### Technical Requirements

- [ ] Create `native/pkg/mcp/tools/sign_message_tool.go`
- [ ] Implement EIP-191 personal message signing
- [ ] Implement EIP-712 typed data signing
- [ ] Add message validation and sanitization
- [ ] Support signature verification for testing
- [ ] Integrate with existing wallet manager

### Security Requirements

- [ ] Validate message content for safety
- [ ] Implement signing rate limits
- [ ] Add warnings for potentially dangerous messages
- [ ] Prevent signing of transaction-like messages
- [ ] Log all signing operations for audit
- [ ] Support message preview before signing

## Acceptance Criteria

- [ ] Tool can sign personal messages (EIP-191)
- [ ] Tool can sign typed data (EIP-712)
- [ ] Signatures are verifiable on-chain and off-chain
- [ ] Security validations prevent dangerous operations
- [ ] Integration tests pass with real dApp interactions
- [ ] Performance requirements met

## Implementation Details

### Files to Create/Modify

- `native/pkg/mcp/tools/sign_message_tool.go` (new)
- `native/pkg/wallet/signing.go` (new)
- `native/pkg/wallet/eip712.go` (new)
- `native/pkg/security/message_validator.go` (extend)
- `native/cmd/main.go` (register tool)

### API Schema

```json
{
  "name": "sign_message",
  "description": "Sign a message with the wallet private key",
  "inputSchema": {
    "type": "object",
    "properties": {
      "message": { "type": "string", "description": "Message to sign" },
      "type": { "type": "string", "enum": ["personal", "typed_data"], "default": "personal" },
      "typed_data": {
        "type": "object",
        "description": "EIP-712 typed data (required if type is typed_data)"
      },
      "format": { "type": "string", "enum": ["hex", "base64", "raw"], "default": "hex" },
      "chain": {
        "type": "string",
        "enum": ["ethereum", "bsc"],
        "description": "Chain for EIP-712 domain"
      }
    },
    "required": ["message"]
  }
}
```

### Response Format

```json
{
  "signature": "0x...",
  "message_hash": "0x...",
  "address": "0x...",
  "type": "personal",
  "format": "hex",
  "timestamp": "2025-06-24T07:00:00Z"
}
```

### EIP-712 Support

```json
{
  "types": {
    "EIP712Domain": [
      { "name": "name", "type": "string" },
      { "name": "version", "type": "string" },
      { "name": "chainId", "type": "uint256" },
      { "name": "verifyingContract", "type": "address" }
    ],
    "Person": [
      { "name": "name", "type": "string" },
      { "name": "wallet", "type": "address" }
    ]
  },
  "primaryType": "Person",
  "domain": {
    "name": "Algonius",
    "version": "1",
    "chainId": 1,
    "verifyingContract": "0x..."
  },
  "message": {
    "name": "Alice",
    "wallet": "0x..."
  }
}
```

## Dependencies

- Requires go-ethereum library for signing operations
- Related to wallet manager and security modules
- May integrate with hardware wallet support in future

## Testing Requirements

- [ ] Unit tests for message signing
- [ ] EIP-191 compliance tests
- [ ] EIP-712 compliance tests
- [ ] Signature verification tests
- [ ] Security validation tests

## Security Considerations

- [ ] **Message Content Validation**: Prevent signing of transaction-like data
- [ ] **Rate Limiting**: Limit signing operations per time period
- [ ] **Audit Logging**: Record all signing operations
- [ ] **User Confirmation**: Optional confirmation for sensitive messages
- [ ] **Domain Validation**: Validate EIP-712 domain parameters

## Use Cases

- **dApp Authentication**: Prove wallet ownership to decentralized applications
- **Off-chain Authorization**: Sign permits and approvals
- **Message Verification**: Verify message authenticity
- **Cross-chain Communication**: Sign messages for bridge protocols

## Configuration

- [ ] Signing rate limits
- [ ] Dangerous message patterns
- [ ] Audit log settings
- [ ] Domain whitelist for EIP-712

## References

- EIP-191: https://eips.ethereum.org/EIPS/eip-191
- EIP-712: https://eips.ethereum.org/EIPS/eip-712
- Technical Spec: `docs/teck_spec.md`
- MCP API Documentation: `docs/apis/native_host_mcp_api.md`
