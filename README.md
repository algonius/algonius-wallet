# Algonius Wallet

A secure browser extension wallet with native host integration.

## Prerequisites

- Go 1.20+
- Node.js 16+
- Chrome or Firefox browser
- GNU Make

## Getting Started

### 1. Build and Run Native Host

```bash
cd projects/algonius-wallet
make build  # Build the native host
make run    # Run with 30s timeout
```

### 2. Install System-wide (Optional)

```bash
make install  # Install to /usr/local/bin
```

### 3. Start Browser Extension MCP Server

1. Load the extension in developer mode
2. The extension will automatically start MCP server on port 8080

### 4. Connect Native Host to MCP Server

The native host will automatically connect when both are running.

## Development

### Makefile Targets

- `make build`: Build native host
- `make install`: Install system-wide
- `make run`: Run with timeout

### Native Host Features

- Wallet creation and management
- Secure key storage
- Transaction signing
- MCP server integration

### Extension Features

- UI for wallet operations
- MCP server for native communication
- Transaction confirmation flow

## Troubleshooting

- If connection fails, ensure MCP server is running first
- Check firewall settings for port 8080
- Use `make run` for automatic timeout

## License

MIT
