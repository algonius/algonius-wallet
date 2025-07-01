# Native Host Application组件详细设计 (安全增强版)

## 1. 组件内模块图

```
┌──────────────────────────────────────────────────────┐
│             Native Host App (Go)                     │
│  ┌─────────────┐ ┌─────────────────┐ ┌─────────────┐ │
│  │ MCP Server  │ │ Wallet Manager  │ │ Event Broad-│ │
│  │ - Tools     │ │ - Multi-chain   │ │ caster      │ │
│  │ - Resources │ │ - Private Keys  │ │ - Observer  │ │
│  └─────────────┘ └─────────────────┘ └─────────────┘ │
│                      ┌────────────────────────┐      │
│                      │ Native Messaging       │      │
│                      │ - Chrome Integration   │      │
│                      └────────────────────────┘      │
└──────────────────────────────────────────────────────┘
```

## 2. 每个模块的详细设计

### 2.1 MCP Server模块

#### 模块描述

MCP Server模块负责实现Model Context Protocol协议，提供AI Agent交互的工具和资源接口。该模块是Native Host与AI Agent通信的核心组件，现在直接集成了事件广播功能，并移除了敏感的import_wallet工具。

#### 模块职责

1. 接收和验证来自AI Agent的MCP协议请求
2. 将请求路由到相应的处理函数
3. 调用Wallet Manager执行实际的钱包操作
4. 通过Event Broadcaster发送实时事件
5. 管理MCP客户端会话
6. 提供标准化的错误处理和响应格式
7. 实现SSE事件流支持

#### 模块接口设计

```go
// pkg/mcp/tools.go
type MCPTools struct {
    CreateWallet       Tool `json:"create_wallet"`
    GetBalance         Tool `json:"get_balance"`
    SendTransaction    Tool `json:"send_transaction"`
    ConfirmTransaction Tool `json:"confirm_transaction"`
    GetTransactions    Tool `json:"get_transactions"`
    SignMessage        Tool `json:"sign_message"`
    SwapTokens         Tool `json:"swap_tokens"`
}

// pkg/mcp/server.go
type MCPServer struct {
    tools map[string]ToolHandler
    walletManager *wallet.Manager
    eventBroadcaster *EventBroadcaster
}

func (s *MCPServer) Initialize() error
func (s *MCPServer) CallTool(method string, params map[string]interface{}) (*ToolResult, error)
func (s *MCPServer) HandleGetBalance(params map[string]interface{}) (*ToolResult, error)
func (s *MCPServer) HandleSendTransaction(params map[string]interface{}) (*ToolResult, error)
func (s *MCPServer) HandleConfirmTransaction(params map[string]interface{}) (*ToolResult, error)
func (s *MCPServer) HandleSignMessage(params map[string]interface{}) (*ToolResult, error)
func (s *MCPServer) HandleSwapTokens(params map[string]interface{}) (*ToolResult, error)

// 新增SSE相关接口
func (s *MCPServer) StartEventStream(clientID string) (<-chan *Event, error)
func (s *MCPServer) HandleEventSubscription() http.HandlerFunc
func (s *MCPServer) CallToolWithTimeout(method string, params map[string]interface{}, timeout time.Duration) (*ToolResult, error)

// 新增安全相关的接口
func (s *MCPServer) IsWalletReady(address string) bool
func (s *MCPServer) GetWalletPublicKey(address string) (string, error)
func (s *MCPServer) GetSupportedChains() []string
func (s *MCPServer) GetSupportedTokens(chain string) []string
func (s *MCPServer) GetWalletStatus(address string) WalletStatus
```

#### 核心功能时序图

```
┌──────────┐    ┌────────────┐    ┌────────────┐
│ AI Agent │───►│  MCP Server│───►│Wallet Mgr  │
└──────────┘    └────────────┘    └────────────┘
     │                │                │
     │ get_balance()  │                │
     │───────────────►│                │
     │                │ queryBalance() │
     │                │────────────────►│
     │                │                ◄│
     │                ◄─────────────────│
     ◄───────────────────────────────────│
```

#### 文字描述

MCP Server模块接收来自AI Agent的工具调用请求，解析请求参数，调用相应的业务逻辑处理函数，并返回格式化的结果。对于需要区块链交互的操作，会调用Wallet Manager模块进行实际的链上操作。现在直接支持SSE事件流，无需通过HTTP Server模块。

---

#### 【MCP 工具与资源接口一览（开发实现用）】

详见 docs/apis/native_host_mcp_api.md

---

### 2.2 Wallet Manager模块

#### 模块描述

Wallet Manager模块负责多链钱包管理，包括私钥生成和存储、交易签名以及与区块链网络的交互。

#### 模块接口设计

```go
// pkg/wallet/manager.go
type Manager struct {
    wallets map[string]*Wallet
    chains  map[string]Blockchain
}

func (m *Manager) NewWallet() (*Wallet, error)
func (m *Manager) ImportWallet(mnemonic string) (*Wallet, error)
func (m *Manager) GetBalance(address, token string) (*big.Int, error)
func (m *Manager) SendTransaction(from, to, value string, options ...TxOption) (*Transaction, error)
func (m *Manager) SignTransaction(tx *Transaction) ([]byte, error)
func (m *Manager) GetWallet(address string) (*Wallet, error)
func (m *Manager) IsWalletReady(address string) bool
func (m *Manager) GetWalletStatus(address string) WalletStatus

// pkg/wallet/transaction.go
type Transaction struct {
    Hash      string
    From      string
    To        string
    Value     string
    GasLimit  uint64
    GasPrice  *big.Int
    Nonce     uint64
    Status    string
}

type TxOption func(*Transaction)
func WithGasLimit(limit uint64) TxOption
func WithPriorityFee(fee float64) TxOption

// 钱包状态定义
type WalletStatus struct {
    Address string
    PublicKey string
    Ready bool
    Chains map[string]bool
    LastUsed int64
}
```

#### 核心功能时序图

```
┌────────────┐    ┌────────────┐    ┌────────────┐
│MCP Server  │───►│Wallet Mgr  │───►│ Blockchain │
└────────────┘    └────────────┘    └────────────┘
     │                │                │
     │ sendTransaction│                │
     │───────────────►│                │
     │                │ buildTx()     │
     │                │────────────────►│
     │                │ signTx()      │
     │                │────────────────►│
     │                │ broadcastTx() │
     │                │────────────────►│
     │                ◄─────────────────│
     ◄───────────────────────────────────│
```

#### 文字描述

Wallet Manager模块处理所有与钱包相关的操作，包括创建钱包、查询余额和发送交易。对于链上操作，会调用相应的区块链客户端进行实际的网络交互。敏感操作如钱包导入始终在Native Host内部完成，不会暴露给AI Agent。

### 2.3 Event Broadcaster模块

#### 模块描述

Event Broadcaster模块负责管理事件广播，采用观察者模式将事件推送给所有连接的AI Agent。现在直接与MCP Server集成，管理MCP客户端的事件订阅。

#### 模块接口设计

```go
// pkg/api/events.go
type EventBroadcaster struct {
    clients map[string]chan *Event
    mcpClients map[string]chan *Event
    mu      sync.RWMutex
}

type Event struct {
    Type string      `json:"type"`
    Data interface{} `json:"data"`
    Timestamp int64  `json:"timestamp"`
}

func NewEventBroadcaster() *EventBroadcaster
func (eb *EventBroadcaster) Subscribe(clientID string) chan *Event
func (eb *EventBroadcaster) Unsubscribe(clientID string)
func (eb *EventBroadcaster) Broadcast(event *Event)

// 新增MCP客户端管理接口
func (eb *EventBroadcaster) SubscribeMCPClient(clientID string) chan *Event
func (eb *EventBroadcaster) BroadcastToMCP(clientID string, event *Event)
func (eb *EventBroadcaster) BroadcastEventToAll(event *Event)
```

#### 核心功能时序图

```
┌────────────┐    ┌────────────┐    ┌────────────┐
│Event Source│───►│Event Broad-│───►│AI Agent    │
│(Transaction│    │caster      │    │(SSE Client)│
│Confirmation│    └────────────┘    └────────────┘
│, Balance   │          │                │
│Update, etc │          │                │
└────────────┘          │                │
                        │ Broadcast(event)│
                        │────────────────►│
                        │                ◄│
                        ◄─────────────────│
```

#### 文字描述

Event Broadcaster模块维护所有连接的AI Agent客户端，当有新事件产生时(如交易确认、余额更新等)，会将事件广播给所有订阅的AI Agent。客户端通过MCP协议的SSE功能接收实时事件。

### 2.4 Native Messaging模块

#### 模块描述

Native Messaging模块负责实现与浏览器扩展的通信，处理来自扩展的消息并返回响应。现在增加了对MCP协议的代理支持，可以将浏览器扩展作为MCP客户端代理。

#### 模块接口设计

```go
// pkg/messaging/native.go
type NativeMessaging struct {
    wallet    *wallet.Manager
    mcpServer *MCPServer
}

func (nm *NativeMessaging) HandleMessage(msg map[string]interface{}) map[string]interface{}
func (nm *NativeMessaging) handleConnectWallet(msg map[string]interface{}) map[string]interface{}
func (nm *NativeMessaging) handleSignTransaction(msg map[string]interface{}) map[string]interface{}
func (nm *NativeMessaging) handleGetAccounts() map[string]interface{}

// 新增的钱包导入接口
func (nm *NativeMessaging) handleImportWallet(msg map[string]interface{}) map[string]interface{}
func (nm *NativeMessaging) handleExportWallet(msg map[string]interface{}) map[string]interface{}
func (nm *NativeMessaging) handleBackupWallet(msg map[string]interface{}) map[string]interface{}
func (nm *NativeMessaging) handleRestoreWallet(msg map[string]interface{}) map[string]interface{}
func (nm *NativeMessaging) handleSetPassword(msg map[string]interface{}) map[string]interface{}

// 新增事件处理接口
func (nm *NativeMessaging) handleEventSubscription(msg map[string]interface{}) map[string]interface{}
func (nm *NativeMessaging) sendEventToExtension(event *Event)
```

#### 核心功能时序图

```
┌────────────┐    ┌────────────┐    ┌────────────┐
│Browser Ext │───►│Native Msg  │───►│Wallet Mgr  │
│(Extension) │    │(Host)      │    │             │
└────────────┘    └────────────┘    └────────────┘
     │                │                │
     │ import_wallet  │                │
     │────────────────►│                │
     │                │ validateInput() │
     │                │────────────────►│
     │                │ storeWallet()  │
     │                │────────────────►│
     │                │                ◄│
     │                ◄─────────────────│
     ◄───────────────────────────────────│
```

#### 文字描述

Native Messaging模块接收来自浏览器扩展的消息，解析消息类型，调用相应的处理函数。对于需要钱包操作的消息，会转发给Wallet Manager模块处理。对于事件订阅请求，会通过MCP Server模块建立SSE连接，将事件推送到浏览器扩展。钱包导入等敏感操作通过此模块处理，不暴露给AI Agent。

## 3. 安全增强后的模块交互

```
┌──────────┐    ┌────────────┐    ┌────────────┐
│ AI Agent │───►│  MCP Server│───►│Wallet Mgr  │
└──────────┘    └────────────┘    └────────────┘
     │                │                │
     │                │                │
     │    ┌──────────▼──────────┐     │
     │    │ Event Broadcaster   │◄────┘
     │    │ (SSE)               │
     │    └──────────┬──────────┘
     │               │
┌──────────┐    ┌────────────────────────┐
│ Browser │───►│ Native Messaging (Host)│
│Extension│    │                        │
└──────────┘    └────────────────────────┘
     │                │
     │ import_wallet  │
     │────────────────►│
     │                │
     │                │
     │                │
     │                │
     │                │
     │                │
     │                │
     │                │
     │                │
     │                │
     ◄────────────────┘
```

## 4. 核心交互流程

### 4.1 DEX交易确认流程 (MCP+SSE)

```
┌──────────┐    ┌────────────┐    ┌────────────┐
│ DEX网站  │───►│ 浏览器扩展 │───►│ Native Host│
└──────────┘    └────────────┘    └────────────┘
     ▲                 │                  │
     │                 │                  │
     │                 │                  │
┌──────────┐    ┌────────────┐    ┌────────────┐
│AI Agent  │◄───│ Native Host│◄───│MCP Server  │
└──────────┘ SSE  └────────────┘    └────────────┘
     │                │                │
     │                │                │
     │                │                │
┌──────────┐    ┌────────────────────────┐
│ Browser │◄───│ Native Messaging       │
│Extension│    │ (Host)                │
└──────────┘    └────────────────────────┘
```

### 4.2 余额查询流程 (MCP直接交互)

```
┌──────────┐    ┌────────────┐    ┌────────────┐
│AI Agent  │───►│  MCP Server│───►│Wallet Mgr  │
└──────────┘    └────────────┘    └────────────┘
     │                │                │
     │ get_balance()  │                │
     │───────────────►│                │
     │                │ queryBalance() │
     │                │────────────────►│
     │                │                ◄│
     │                ◄─────────────────│
     ◄───────────────────────────────────│
```

### 4.3 钱包导入流程 (通过浏览器扩展)

```
┌────────────┐    ┌────────────────────────┐    ┌────────────────────┐
│ Browser UI │──1─►│ 浏览器扩展 │──2─►│ Native Host│──3─►│区块链网络│
└────────────┘      └────────────────────────┘    └────────────────────┘
      ▲                   │                │
      │                   │                │
      └─────────4─────────┴───────5───────┘
                                 6
```

1. **用户界面**: 用户通过浏览器扩展的安全界面输入助记词和密码
2. **扩展验证**: 浏览器扩展验证助记词格式和密码强度
3. **本地存储**: Native Host在本地安全存储加密后的钱包信息
4. **用户确认**: 用户通过扩展界面确认导入操作
5. **消息转发**: 浏览器扩展通过Native Messaging发送加密的导入请求
6. **链上验证**: Native Host验证钱包地址并存储相关信息

### 4.4 直接转账流程 (MCP直接交互)

```
┌──────────┐    ┌────────────┐    ┌────────────┐
│AI Agent  │───►│  MCP Server│───►│Wallet Mgr  │
└──────────┘    └────────────┘    └────────────┘
     │                │                │
     │ send_transaction│                │
     │───────────────►│                │
     │                │ buildTx()     │
     │                │────────────────►│
     │                │ signTx()      │
     │                │────────────────►│
     │                │ broadcastTx() │
     │                │────────────────►│
     │                │                ◄│
     │                ◄─────────────────│
     ◄───────────────────────────────────│
```

## 5. 模块间交互

所有模块通过清晰定义的接口进行交互：

1. **AI Agent与MCP Server**: 通过MCP协议进行工具调用和事件订阅
2. **MCP Server与Wallet Manager**: 执行钱包相关操作
3. **Wallet Manager与Blockchain**: 与区块链网络交互
4. **Event Source与Event Broadcaster**: 广播系统事件
5. **Event Broadcaster与AI Agent**: 通过SSE推送事件
6. **Browser Extension与Native Messaging**: 通过Native Messaging API通信
7. **Native Messaging与MCP Server**: 处理浏览器扩展的MCP代理请求
8. **Native Messaging与Wallet Manager**: 处理浏览器扩展的敏感操作请求

这种安全增强后的架构保持了系统的核心功能，同时去除了敏感的import_wallet工具，确保了用户的助记词和私钥永远不会暴露给AI Agent，提供了更清晰的协议栈和更强的安全性。
