# Algonius Wallet Technical Context

## Core Technologies

### Programming Languages

- **Go**: Primary language for Native Host application
- **JavaScript/TypeScript**: Used for Chrome extension components
- **JSON**: Data interchange format for messaging between components

### Architecture Components

- **MCP Server**: Implements Model Context Protocol for AI Agent interaction (never exposes sensitive wallet operations)
- **HTTP Server**: Provides REST API and SSE endpoints for real-time communication (for AI Agent only)
- **Wallet Manager**: Handles multi-chain wallet operations and private key management
- **Chrome Extension**: Bridges DApps and Native Host through content scripts and background workers
- **Event Broadcaster**: Manages real-time event distribution via SSE (for AI Agent only)
- **Native Messaging**: Handles all communication with Browser Extension, including all sensitive wallet operations (import/export/backup)

### Communication Protocols

- **Chrome Native Messaging**: Secure communication between browser extension and Native Host (all event flows and sensitive operations)
- **HTTP/REST**: API communication between AI Agent and Native Host
- **Server-Sent Events (SSE)**: Real-time event notifications from Native Host to AI Agent
- **MCP (Model Context Protocol)**: Standard for AI Agent tool interaction

## Development Setup

### Native Host Setup

```bash
# Build Native Host
cd native-host
go build -o algonius-wallet-host main.go

# Install to system
sudo cp algonius-wallet-host /usr/local/bin/
sudo cp manifest.json /etc/opt/chrome/native-messaging-hosts/com.algonius.wallet.json

# Create user config directory
mkdir -p ~/.algonius-wallet
chmod 700 ~/.algonius-wallet
```

### Chrome Extension Setup

```bash
# Load extension in developer mode
1. Open chrome://extensions/
2. Enable "Developer mode"
3. Click "Load unpacked"
4. Select the browser-wallet-extension directory
```

## Technical Dependencies

### Go Dependencies

- **Gin**: HTTP web framework for API endpoints
- **Ethereum Go Client**: Blockchain interaction for Ethereum and EVM-compatible chains
- **Solana Go Client**: Blockchain interaction for Solana
- **BSC Go Client**: Blockchain interaction for Binance Smart Chain
- **Go Crypto Libraries**: For secure key storage and transaction signing
- **WebSockets**: For blockchain event subscriptions
- **SSE Libraries**: For Server-Sent Events implementation (AI Agent only)
- **Zap**: Structured logging
- **Testify**: Go unit/integration testing

### JavaScript Dependencies

- **Chrome Extension APIs**: For browser integration
- **Web3.js/Ethers.js**: For Ethereum-compatible blockchain interaction
- **Solana Web3.js**: For Solana blockchain interaction
- **EventEmitter**: For event handling in extension components
- **Playwright/Puppeteer**: For E2E UI and DApp automation
- **Mocha/Jest**: For JS unit/E2E testing

## Supported Blockchains

- Ethereum and EVM-compatible chains (BSC, Polygon, etc.)
- Solana
- Additional chains planned for future implementation

## Technical Constraints

### Security Constraints

- Private keys must never be exposed to browser extension or AI Agent
- All transaction signing must occur within Native Host application
- All wallet import/export/backup operations are strictly handled via Native Messaging, never exposed to MCP tools or AI Agent
- All event flows between extension and Native Host use Chrome Native Messaging (not SSE)
- All sensitive operations (wallet import, signing) are strictly UI-gated and never accessible to AI Agent or DApp
- Communication between components must be encrypted
- API endpoints must only listen on localhost for security
- **All code comments must be written in English. This is a strict team convention for all source files and documentation code blocks.**

### Performance Constraints

- Transaction signing must complete within 5 seconds to meet user expectations
- Balance updates must reflect within 10 seconds of blockchain confirmation
- SSE event delivery (to AI Agent) must have maximum latency of 2 seconds
- Native Host must handle concurrent requests from multiple AI Agents

### Browser Extension Constraints

- Manifest V3 compatibility required for Chrome Web Store listing
- Content Security Policy restrictions must be observed
- Background service worker lifetime limitations must be managed
- Memory usage must be minimized for background operation

## Tool Usage Patterns

### MCP Tools

The Native Host exposes the following MCP tools for AI Agent interaction (no import_wallet; all import/export/backup is via Native Messaging only):

```go
type MCPTools struct {
    CreateWallet       Tool `json:"create_wallet"`
    GetBalance         Tool `json:"get_balance"`
    SendTransaction    Tool `json:"send_transaction"`
    ConfirmTransaction Tool `json:"confirm_transaction"`
    GetTransactions    Tool `json:"get_transactions"`
    SignMessage        Tool `json:"sign_message"`
    SwapTokens         Tool `json:"swap_tokens"`
}
```

### API Endpoints

The HTTP server provides the following key endpoints:

```
# MCP endpoints
POST /mcp/initialize - Initialize MCP connection
POST /mcp/call - Call MCP tool with parameters
GET /mcp/events - SSE endpoint for real-time events (AI Agent only)

# Wallet API
GET /api/wallets - List all wallets
POST /api/wallets - Create new wallet
GET /api/wallets/:id/balance - Get wallet balance
POST /api/transactions - Send transaction
```

### Event Types

The system broadcasts the following event types via SSE (to AI Agent only):

- `transaction_confirmation_needed`: AI Agent needs to approve/reject transaction
- `transaction_confirmed`: Transaction confirmed on blockchain
- `transaction_error`: Transaction failed
- `balance_updated`: Wallet balance changed
- `connected`: Initial connection confirmation

## E2E Automation Toolchain

- **Playwright/Puppeteer**: Automated DApp and UI interaction
- **Go test**: Native Host API and integration testing
- **Mocha/Jest**: JS unit and E2E testing
- **Mock AI Agent**: For simulating auto/manual decision branches
- **Regression and data cleanup scripts**: For test repeatability

## Additional Notes

- All event flows between Browser Extension and Native Host use Chrome Native Messaging, not SSE.
- All sensitive wallet operations (import, export, backup, restore, set password) are strictly UI-gated and handled via Native Messaging, never exposed to MCP tools or AI Agent.
- The system is designed for extensibility, security, and strict separation of authority between AI Agent, Browser Extension, and Native Host.
