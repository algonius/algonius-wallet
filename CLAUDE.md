# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Architecture

Algonius Wallet is a multi-chain crypto wallet system with three main components:

1. **Browser Extension** (TypeScript/React) - Web3 provider injection, DApp interaction, and UI
2. **Native Host** (Go) - Secure wallet management, transaction signing, and MCP server
3. **AI Agent Communication** - MCP protocol for automated trading and intelligent decision-making

The Native Host serves as the security core, managing private keys and blockchain operations, while the Browser Extension provides the user interface and Web3 compatibility. AI agents interact with the Native Host through MCP (Model Context Protocol) tools and resources.

## Development Commands

### Browser Extension (TypeScript/React)
```bash
# Development
npm run dev              # Start Vite development server
npm run build            # Build extension for production
npm run zip              # Create distribution zip file

# Testing and Quality
npm run test             # Run Vitest tests
npm run test:ci          # Run tests in CI mode
npm run lint             # Run ESLint
npm run format           # Format code with Prettier
```

### Native Host (Go)
```bash
cd native

# Development
make build               # Build native host binary
make run                 # Run native host directly
make install             # Install to ~/.algonius-wallet/bin/

# Testing
make test                # Run all tests (unit + integration)
make unit-test           # Run unit tests only
make integration-test    # Run integration tests only
make lint                # Run golangci-lint (requires installation)

# Cleanup
make clean               # Remove built binaries
make uninstall           # Remove installed binaries
```

## Key Architecture Components

### Browser Extension Structure
- **Background Service Worker** (`src/background/background.ts`) - Main extension controller, manages MCP Host connection
- **Content Scripts** (`src/content/`) - Web3 provider injection into DApps
- **Popup UI** (`src/popup/`) - React-based user interface for wallet operations
- **MCP Host Manager** (`src/mcp/McpHostManager.ts`) - Handles Native Messaging and RPC communication

### Native Host Structure
- **Main Entry** (`cmd/main.go`) - Dual-mode server (Native Messaging + HTTP MCP)
- **MCP Tools** (`pkg/mcp/tools/`) - Blockchain operations (balance, transactions, swaps)
- **MCP Resources** (`pkg/mcp/resources/`) - Status and chain information
- **Wallet Manager** (`pkg/wallet/`) - Multi-chain wallet operations and key management
- **Native Messaging** (`pkg/messaging/`) - Browser extension communication

### Communication Patterns
- **Browser Extension ↔ Native Host**: Native Messaging API for secure communication
- **AI Agent ↔ Native Host**: MCP protocol over HTTP (port 9444) for tool/resource access
- **DApps ↔ Browser Extension**: Web3 provider (EIP-1193) for blockchain interaction

## Security Architecture

The Native Host is the only component with access to private keys and signing capabilities. The Browser Extension handles UI and Web3 compatibility, while MCP tools provide controlled access for AI agents without exposing sensitive operations like private key import.

### Security Boundaries
- **AI Agent Access**: Only through MCP tools/resources - cannot access wallet import/export
- **Browser Extension Access**: Can perform sensitive operations via Native Messaging
- **Private Key Protection**: Keys never leave Native Host, encrypted with AES-256-GCM
- **User Control**: All transactions require user confirmation through Popup UI

## MCP Tools & Resources

The Native Host exposes the following tools for AI agents:

### MCP Tools
- `create_wallet` - Create new wallet (requires user authorization)
- `get_balance` - Query wallet balance
- `send_transaction` - Send blockchain transaction (requires user authorization)
- `confirm_transaction` - Query transaction status
- `get_transactions` - Get wallet transaction history
- `sign_message` - Sign message with wallet (requires user authorization)
- `swap_tokens` - Token swap functionality (requires user authorization)

### MCP Resources
- `wallet_status` - Current wallet status information
- `supported_chains` - List of supported blockchain networks
- `supported_tokens` - Supported tokens for each chain

## Native Messaging API

The Browser Extension can access these additional RPC methods through Native Messaging:

- `import_wallet` - Import wallet from mnemonic (extension-only)
- `export_wallet` - Export wallet data (extension-only)
- `get_wallet_info` - Get comprehensive wallet information
- `send_transaction` - Send transactions with full UI control

## Supported Blockchains

- **Ethereum** (`"ethereum"`, `"eth"`)
- **Binance Smart Chain** (`"bsc"`, `"binance"`)

Both chains use Ethereum-compatible addressing and are interoperable.

## Build Configuration

- **Extension Build**: Vite with React plugin, outputs to `dist/` directory
- **Native Host Build**: Go modules with specific binary output to `bin/` directory
- **Testing**: Vitest for TypeScript, Go testing framework for Native Host

## Integration Points

- **MCP Server**: Runs on port 9444 (configurable via SSE_PORT)
- **Native Messaging**: Uses Chrome's native messaging API with manifest registration
- **Web3 Provider**: Injected into DApp pages for blockchain interaction
- **Event Broadcasting**: SSE events for real-time updates to AI agents

## Development Notes

When working on this codebase:
- The Native Host must be built and running for full functionality
- Browser extension development requires loading as unpacked extension
- MCP tools are the primary interface for AI agent interactions
- All sensitive operations (key import, signing) are handled by Native Host
- Test files are co-located with source code (`.test.ts`, `_test.go`)
- E2E tests require Chrome, Ganache/Hardhat, and test DApp setup

## Error Handling

The system uses structured error codes:
- Native Messaging errors: -32001 to -32099 range
- MCP tool errors: Standard MCP error format
- All sensitive operations log errors without exposing private data

## Performance Considerations

- Native Host runs dual HTTP/Native Messaging servers
- Event broadcasting uses observer pattern for efficient updates
- Wallet operations are optimized for multi-chain support
- RPC timeout: 5 seconds default, configurable per request