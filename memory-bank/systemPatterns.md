# Algonius Wallet System Patterns

## Architecture Overview

Algonius Wallet implements a three-tier architecture with clear separation of concerns:

```
┌─────────────────┐                    ┌─────────────────┐
│   AI Agent      │                    │  Browser Ext    │
│ (Claude/GPT)    │◄─── MCP/HTTP ────► │                 │
│  - Existing     │     :9444/mcp      │                 │
│  - Cline/SSE    │◄─── MCP/SSE ──────► │                 │
└─────────────────┘     :9444/mcp/sse  └─────────┬───────┘
         │                                       │
         │ Unified HTTP/SSE                      │ Native
         │ Port :9444                            │ Messaging
         ▼                                       ▼
┌──────────────────────────────────────────────────────┐
│             Native Host App (Go)                     │
│  ┌──────────────────┐ ┌─────────────┐ ┌─────────────┐ │
│  │ Unified MCP      │ │ HTTP Server │ │ Wallet Mgr  │ │
│  │ - Streamable HTTP│ │ - REST API  │ │ Multi-chain │ │
│  │ - Pure SSE       │ │ - SSE       │ │ Private Keys│ │
│  │ - Dual Transport │ │ - Unified   │ │ - Security  │ │
│  └──────────────────┘ └─────────────┘ └─────────────┘ │
└──────────────────────────────────────────────────────┘
```

## Component Relationships

### 1. Native Host Components

- **Unified MCP Server**: Exposes tools and resources via dual transport protocols
  - **Streamable HTTP**: Compatible with existing MCP clients (`/mcp`)
  - **Pure SSE**: Compatible with SSE-only clients like Cline (`/mcp/sse`, `/mcp/message`)
  - **Transport Agnostic**: Same MCP server instance serves both protocols
  - **Security**: Never exposes sensitive wallet operations via either transport
- **HTTP Server**: Provides REST API and unified transport endpoints (for AI Agent only)
- **Wallet Manager**: Handles multi-chain wallet operations and private key management
- **Event Broadcaster**: Manages real-time event distribution via SSE (for AI Agent only)
- **Native Messaging**: Handles all communication with Browser Extension, including all sensitive wallet operations (import/export/backup)

### 2. Browser Extension Components

- **Background Service Worker**: Manages Native Host connection and message routing (via Native Messaging only)
- **Content Script**: Injects Web3 provider into web pages and handles DApp communication
- **Popup Interface**: Provides user interface for wallet management and settings; all sensitive operations are strictly UI-gated

## Core Data Flow Patterns

### DEX Transaction Confirmation Flow

```
┌──────────┐      ┌───────────┐      ┌────────────┐      ┌────────┐
│ DEX网站  │──1──►│ 浏览器扩展 │──2──►│ Native Host│──3──►│AI Agent│
└──────────┘      └───────────┘      └────────────┘      └────────┘
      ▲                 │                  ▲                 │
      │                 │                  │                 │
      └────────7────────┴───────6─────────┴────────5────────┘
                                 4
```

1. DEX website calls `window.ethereum.request({method: 'eth_sendTransaction', params: [txParams]})`
2. Content Script receives request, Background Service Worker forwards via Native Messaging
3. Native Host sends `transaction_confirmation_needed` event to AI Agent via SSE
4. AI Agent analyzes transaction parameters and makes decision
5. AI Agent calls `confirm_transaction` MCP tool with decision
6. Native Host signs transaction and returns to extension
7. Extension returns signed transaction to DEX website for blockchain submission

### Balance Query Flow

```
┌────────┐                  ┌────────────┐                ┌─────────┐
│AI Agent│─────HTTP/MCP────►│Native Host │────RPC/API────►│区块链网络│
└────────┘                  └────────────┘                └─────────┘
    │                             │                            │
    │ 1. get_balance工具          │ 2. 查询链上数据             │
    │                             │                            │
    │ 3. 返回余额 + SSE更新       │ 4. 监听余额变化             │
    │◄────────────────────────────┤◄───────────────────────────┘
```

### Direct Transfer Flow

```
┌────────┐                  ┌────────────┐                ┌─────────┐
│AI Agent│─────HTTP/MCP────►│Native Host │────RPC/API────►│区块链网络│
└────────┘                  └────────────┘                └─────────┘
    │                             │                            │
    │ 1. send_transaction         │ 2. 构建&签名交易            │
    │                             │                            │
    │                             │ 3. 广播交易                │
    │                             │                            │
    │ 4. 返回交易哈希             │ 5. 监听确认状态             │
    │◄────────────────────────────┤◄───────────────────────────┘
    │                             │
    │ 6. SSE推送确认状态          │
    │◄────────────────────────────┘
```

## Security Boundaries

- **AI Agent**: Can only access controlled MCP tools/resources (create_wallet, get_balance, send_transaction, confirm_transaction, get_transactions, sign_message, swap_tokens, wallet_status, supported_chains, supported_tokens). Never receives or triggers sensitive operations (wallet import/export/backup).
- **Browser Extension**: All sensitive wallet operations (import, export, backup, restore, set password) are strictly handled via Native Messaging and UI, never exposed to MCP tools or AI Agent.
- **Native Host**: Enforces strict separation between MCP and Native Messaging flows. All event flows between extension and Native Host use Native Messaging (not SSE). SSE is only used for AI Agent event push.

## Technical Decision Records

### Native Host Dependency Management: Parameter Dependency Injection

**Decision**: All dependencies in the native-host (Go) component are managed via parameter (constructor) dependency injection.

**Rationale**:

- Promotes explicit, testable, and modular code.
- Avoids global variables and hidden dependencies.
- Facilitates unit testing and mocking.
- Aligns with modern Go best practices for scalable backend systems.

**Pattern**:

- All service, manager, and utility dependencies are passed as parameters to constructors (struct initializers).
- No package-level singletons or implicit imports for core business logic.
- Example:

  ```go
  type WalletManager struct {
      Logger   Logger
      Database DB
  }

  func NewWalletManager(logger Logger, db DB) *WalletManager {
      return &WalletManager{
          Logger: logger,
          Database: db,
      }
  }
  ```

### 1. Chrome Native Messaging vs. Web Extension Messaging

**Decision**: Use Chrome Native Messaging API for communication between browser extension and Native Host

**Rationale**:

- Provides secure, bidirectional communication channel
- Supports long-lived connections necessary for real-time updates
- Allows Native Host application to run with elevated privileges for secure key storage
- Bypasses web extension content security policies for blockchain operations

### 2. SSE vs. WebSockets for AI Agent Communication

**Decision**: Use Server-Sent Events (SSE) for real-time notifications from Native Host to AI Agent

**Rationale**:

- Simpler implementation than WebSockets for one-way real-time communication
- Built-in reconnection mechanisms
- Text-based protocol simplifies debugging
- Lower overhead than WebSockets for unidirectional event streaming
- HTTP-based, easier to integrate with existing web infrastructure

### 3. Multi-chain Support Architecture

**Decision**: Implement wallet functionality using a modular, chain-agnostic interface with specific implementations for each blockchain

**Rationale**:

- Allows for clean separation of chain-specific logic
- Enables adding new blockchains without major architecture changes
- Facilitates consistent API for AI Agent interaction across chains
- Simplifies testing of individual chain implementations

### 4. Transaction Signing Approach

**Decision**: Perform all transaction signing within the Native Host application, never exposing private keys to the browser extension or AI Agent

**Rationale**:

- Minimizes attack surface for private key exposure
- Leverages operating system security features for key storage
- Allows for implementation of additional security measures (hardware wallet integration, etc.)
- Maintains single source of truth for transaction validation logic

## Critical Implementation Paths

### Unified MCP Server Implementation Pattern

**All code comments must be written in English. This is a strict team convention for all source files and documentation code blocks.**

The unified MCP server follows a dual-transport architecture pattern:

```go
// Unified server setup in main.go
func setupUnifiedMCPServer(mcpServer *server.MCPServer, port string) *http.Server {
    mux := http.NewServeMux()
    
    // Streamable HTTP - compatible with existing clients
    streamableServer := server.NewStreamableHTTPServer(mcpServer, 
        server.WithEndpointPath("/mcp"))
    mux.Handle("/mcp", streamableServer)
    
    // Pure SSE - compatible with Cline and other SSE-only clients
    sseServer := server.NewSSEServer(mcpServer,
        server.WithStaticBasePath("/mcp"),
        server.WithSSEEndpoint("/sse"),
        server.WithMessageEndpoint("/message"),
        server.WithUseFullURLForMessageEndpoint(false))
    mux.Handle("/mcp/sse", sseServer.SSEHandler())
    mux.Handle("/mcp/message", sseServer.MessageHandler())
    
    return &http.Server{Addr: port, Handler: mux}
}
```

### MCP Tool Implementation Pattern

MCP tools follow a consistent implementation pattern across both transports:

```go
// Tool definition
type MCPTools struct {
    CreateWallet       Tool `json:"create_wallet"`
    GetBalance         Tool `json:"get_balance"`
    SendTransaction    Tool `json:"send_transaction"`
    ConfirmTransaction Tool `json:"confirm_transaction"`
    GetTransactions    Tool `json:"get_transactions"`
    SignMessage        Tool `json:"sign_message"`
    SwapTokens         Tool `json:"swap_tokens"`
    // No import_wallet here; all import/export/backup is via Native Messaging only
}

// Tool handler implementation (works with both HTTP and SSE)
func (s *MCPServer) HandleGetBalance(params map[string]interface{}) (*ToolResult, error) {
    // Parameter validation
    // Business logic implementation  
    // Return formatted result (same for both transports)
}
```

### Event Broadcasting Pattern

The EventBroadcaster follows the observer pattern to distribute events to all connected AI Agents (via SSE only):

```go
type EventBroadcaster struct {
    clients map[string]chan *Event
    mu      sync.RWMutex
}

func (eb *EventBroadcaster) Subscribe(clientID string) chan *Event {
    // Add client to subscription list
}

func (eb *EventBroadcaster) Broadcast(event *Event) {
    // Send event to all subscribed clients
}
```

## Additional Notes

### Transport Layer Compatibility

- **Dual Transport Support**: Native Host supports both HTTP Streamable and Pure SSE transports simultaneously
- **Endpoint Mapping**:
  - `/mcp` → Streamable HTTP (existing clients, backward compatible)
  - `/mcp/sse` → Pure SSE connection (Cline, SSE-only clients)  
  - `/mcp/message` → SSE message endpoint for communication
- **Client Testing**: Both transports tested with identical functionality and results
- **Official SSE Integration**: Uses `mark3labs/mcp-go/client/sse.go` for reliable SSE client implementation

### Security & Communication Boundaries

- All event flows between Browser Extension and Native Host use Chrome Native Messaging, not SSE.
- All sensitive wallet operations (import, export, backup, restore, set password) are strictly UI-gated and handled via Native Messaging, never exposed to MCP tools or AI Agent.
- The system is designed for extensibility, security, and strict separation of authority between AI Agent, Browser Extension, and Native Host.
- **GitHub Issue #4 RESOLVED**: SSE transport layer compatibility fully implemented and tested.
