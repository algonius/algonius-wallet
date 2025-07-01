# 浏览器扩展组件详细设计

## 1. 组件内模块图

```
┌──────────────────────────────────────────────────────┐
│             浏览器扩展 (Chrome)                    │
│  ┌─────────────────┐ ┌──────────────────┐ ┌─────────┐ │
│  │ Background      │ │ Content Script   │ │ Popup │ │
│  │ Service Worker  │ │ - Web3注入       │ │ UI    │ │
│  │ - Native Host  │ │ - DEX检测         │ │ - 状态│ │
│  │   连接          │ │ - 消息转发        │ │ - 设置│ │
│  └─────────────────┘ └──────────────────┘ └─────────┘ │
│                      ┌────────────────────────┐      │
│                      │ Web3 Provider         │      │
│                      │ - EIP-1193兼容       │      │
│                      └────────────────────────┘      │
└──────────────────────────────────────────────────────┘
```

## 2. 每个模块的详细设计

### 2.1 Background Service Worker模块

#### 模块描述

Background Service Worker模块是浏览器扩展的核心组件，负责管理与Native Host的通信、消息路由和交易处理。所有事件广播和资源变化通知均通过Native Messaging通道完成，不再使用SSE。

#### 模块职责

1. 管理与Native Host的持久连接
2. 处理来自Content Script的消息
3. 路由请求到Native Host
4. 管理交易确认流程
5. 提供自动重连机制
6. 处理Native Messaging事件流（包括资源变化、交易状态、余额更新等事件）

#### 模块接口设计

```javascript
class NativeHostConnection {
  constructor() {
    this.port = null;
    this.pendingRequests = new Map();
    this.pendingTransactions = new Map();
    this.init();
  }

  init() {
    /* 初始化方法 */
  }
  connectToNativeHost() {
    /* 建立与Native Host的连接 */
  }
  setupMessageHandling() {
    /* 设置消息处理 */
  }
  setupContentScriptHandling() {
    /* 设置Content Script消息处理 */
  }

  handleContentScriptMessage(request, sender, sendResponse) {
    /* 处理Content Script消息 */
  }
  waitForTransactionConfirmation(txId) {
    /* 等待交易确认 */
  }
  sendToNativeHost(action, params) {
    /* 发送请求到Native Host */
  }
  handleNativeResponse(response) {
    /* 处理Native Host响应 */
  }
  generateRequestId() {
    /* 生成请求ID */
  }
}

// Chrome扩展消息处理
chrome.runtime.onStartup.addListener(() => {
  /* 启动时连接 */
});
chrome.runtime.onInstalled.addListener(() => {
  /* 安装时连接 */
});
```

#### 核心功能时序图

```
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│DEX Website │───►│ 浏览器扩展 │───►│ Native Host│
└────────────┘    │(Content Script)         │    └────────────┘
     ▲            └────────────────────────┘          │
     │                             │                   │
     │                             │                   │
     │                             │                   │
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│AI Agent   │◄───│ Native Host           │◄───│MCP Server  │
└────────────┘ SSE └────────────────────────┘    └────────────┘
     │                             │                   │
     │                             │                   │
     │    ┌────────────────────┐ │                   │
     │    │ Event Broadcaster │◄─┘                   │
     │    │ (SSE)            │                       │
     │    └────────────────────┘                       │
     │                             │                   │
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│ Browser UI │◄───│ Native Messaging       │◄───│Wallet Mgr  │
│(Popup)     │    │ (Host)                │    │             │
└────────────┘    └────────────────────────┘    └────────────┘
```

#### 文字描述

Background Service Worker模块管理与Native Host的长连接，处理来自Content Script的消息，并协调交易确认流程。它作为Native Host和Content Script之间的消息路由器，确保所有敏感操作通过Native Host完成。对于需要用户交互的操作，会将请求转发给Popup UI。

### 2.2 Content Script模块

#### 模块描述

Content Script模块负责检测网页中的DEX网站，注入Web3提供者，并将Web3请求转发给Background Service Worker。

#### 模块职责

1. 检测当前网页是否为DEX网站
2. 注入Web3提供者
3. 拦截Web3请求
4. 将Web3请求转发给Background Service Worker
5. 处理页面中DApp的Web3调用

#### 模块接口设计

```javascript
class Web3Provider {
  constructor() {
    this.isAIWallet = true;
    this.chainId = '0x1'; // Ethereum mainnet
    this.selectedAddress = null;
    this.setupEventHandlers();
  }

  async request({ method, params }) {
    /* Web3方法实现 */
  }
  setupEventHandlers() {
    /* 设置事件处理 */
  }

  async requestAccounts() {
    /* 请求账户 */
  }
  async sendTransaction(params) {
    /* 发送交易 */
  }
  async personalSign(params) {
    /* 签名消息 */
  }
}

// 注入到页面
if (typeof window.ethereum === 'undefined') {
  window.ethereum = new Web3Provider();
  window.dispatchEvent(new Event('ethereum#initialized'));
}
```

#### 核心功能时序图

```
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│DEX Website │───►│ 浏览器扩展 │───►│ Native Host│
│(Web Page)  │    │(Content Script)         │    └────────────┘
└────────────┘    └────────────────────────┘          │
     │                             │                   │
     │ window.ethereum.request()   │                   │
     │───────────────────────────►│                   │
     │                             │                   │
     │                             │                   │
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│ Browser UI │◄───│ Background Service     │◄───│Wallet Mgr  │
│(Popup)     │ SSE │ Worker (Host)          │    │             │
└────────────┘    └────────────────────────┘    └────────────┘
```

#### 文字描述

Content Script检测当前页面是否为DEX网站，如果是则注入Web3提供者。当DApp调用Web3方法时，Content Script将请求转发给Background Service Worker，由其通过Native Messaging与Native Host通信。所有敏感操作如交易签名、账户访问都在Native Host中完成。

### 2.3 Popup UI模块

#### 模块描述

Popup UI模块提供用户界面，用于钱包管理和设置。

#### 模块职责

1. 显示钱包状态和余额
2. 处理用户交互
3. 提供设置和配置界面
4. 显示交易确认对话框
5. 显示连接的钱包信息
6. 提供账户切换功能

#### 模块接口设计

```javascript
// 钱包状态显示
function displayWalletStatus(status) {
  document.getElementById('wallet-address').textContent = status.address;
  document.getElementById('wallet-balance').textContent = status.balance;
  document.getElementById('chain-status').textContent = status.chain;
}

// 交易确认对话框
function showTransactionConfirmation(transaction) {
  const confirmDialog = document.getElementById('transaction-confirm');
  confirmDialog.style.display = 'block';
  // 显示交易详情
  // 设置批准/拒绝按钮事件处理
}

// 钱包设置界面
function setupSettingsUI() {
  document.getElementById('change-password').addEventListener('click', () => {
    // 处理密码更改
  });

  document.getElementById('backup-wallet').addEventListener('click', () => {
    // 处理钱包备份
  });

  document.getElementById('disconnect-wallet').addEventListener('click', () => {
    // 处理钱包断开
  });
}
```

#### 核心功能时序图

```
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│Browser Ext │───►│ Background Service     │───►│ Native Host│
│(Popup UI)  │    │ Worker (Host)          │    │           │
└────────────┘    └────────────────────────┘    └────────────┘
     │                             │                   │
     │ 用户点击"发送交易"           │                   │
     │───────────────────────────►│                   │
     │                             │                   │
     │                             │                   │
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│ Browser UI │◄───│ Background Service     │◄───│Wallet Mgr  │
│(Popup)     │ SSE │ Worker (Host)          │    │             │
└────────────┘    └────────────────────────┘    └────────────┘
```

#### 文字描述

Popup UI模块提供可视化的钱包管理界面，包括钱包地址显示、余额展示、交易确认对话框和设置界面。所有敏感操作通过Background Service Worker与Native Host交互完成，确保安全边界。

### 2.4 Web3 Provider模块

#### 模块描述

Web3 Provider模块实现了EIP-1193兼容的Web3提供者接口，用于DApp交互。

#### 模块职责

1. 实现EIP-1193兼容的Web3提供者接口
2. 处理DApp的Web3请求
3. 与Background Service Worker通信
4. 简化DApp交互流程
5. 提供标准Web3接口兼容性

#### 模块接口设计

```javascript
class Web3Provider {
  constructor() {
    this.isAIWallet = true;
    this.chainId = '0x1'; // Ethereum mainnet
    this.selectedAddress = null;
    this.setupEventHandlers();
  }

  async request({ method, params }) {
    /* Web3接口实现 */
  }
  setupEventHandlers() {
    /* 设置事件处理 */
  }

  async requestAccounts() {
    /* 请求账户 */
  }
  async sendTransaction(params) {
    /* 发送交易 */
  }
  async personalSign(params) {
    /* 签名消息 */
  }
  async getBalance(address) {
    /* 获取余额 */
  }
  async getChainId() {
    /* 获取链ID */
  }
  async switchChain(chainId) {
    /* 切换链 */
  }
}

// Web3事件处理
window.addEventListener('ethereum#initialized', () => {
  console.log('Web3 provider initialized');
  // 触发钱包连接检查
});
```

#### 核心功能时序图

```
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│DEX Website │───►│ 浏览器扩展 │───►│ Native Host│
│(DApp)     │    │(Content Script)         │    └────────────┘
└────────────┘    └────────────────────────┘          │
     │                             │                   │
     │ window.ethereum.request()   │                   │
     │───────────────────────────►│                   │
     │                             │                   │
     │                             │                   │
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│ Browser UI │◄───│ Background Service     │◄───│Wallet Mgr  │
│(Popup)     │ SSE │ Worker (Host)          │    │             │
└────────────┘    └────────────────────────┘    └────────────┘
```

#### 文字描述

Web3 Provider模块实现了标准的Web3接口，允许DApp与Algonius Wallet交互。当DApp调用Web3方法时，请求通过Content Script转发给Background Service Worker，最终由Native Host处理。用户交互(如交易确认)通过Popup UI展示。

## 3. 模块间交互

```
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│DEX Website │───►│ 浏览器扩展 │───►│ Native Host│
│(DApp)     │    │(Content Script)         │    └────────────┘
└────────────┘    └────────────────────────┘          │
     │                             │                   │
     │ Web3请求                  │                   │
     │───────────────────────────►│                   │
     │                             │                   │
     │                             │                   │
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│ Browser UI │◄───│ Background Service     │◄───│Wallet Mgr  │
│(Popup)     │    │ Worker (Host)          │    │             │
└────────────┘    └────────────────────────┘    └────────────┘
     │                             │                   │
     │ 用户交互                   │                   │
     │───────────────────────────►│                   │
     │                             │                   │
     │                             │                   │
     │    ┌────────────────────┐ │                   │
     │    │ Event Broadcaster │◄─┘                   │
     │    │ (SSE)            │                       │
     │    └────────────────────┘                       │
     │                             │                   │
     │    ┌────────────────────┐ │                   │
     │    │ Native Messaging  │◄─┘                   │
     │    │ (Host)           │                      │
     │    └────────────────────┘                      │
     │                             │                   │
     └─────────────────────────────►│                   │
                                   │                   │
                                   │                   │
                                   │                   │
                                   │                   │
                                   │                   │
                                   │                   │
                                   │                   │
                                   │                   │
                                   │                   │
                                   │                   │
                                   │                   │
```

## 4. 核心交互流程

### 4.1 DEX交易确认流程 (MCP+SSE)

```
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│DEX Website │──1─►│ 浏览器扩展 │──2─►│ Native Host│
│(DApp)     │    │(Content Script)         │    └────────────┘
└────────────┘    └────────────────────────┘          │
     ▲                            │                   │
     │                            │                   │
     │                            │                   │
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│ Browser UI │◄3──│ Background Service     │◄───│MCP Server  │
│(Popup)     │ SSE │ Worker (Host)          │    │             │
└────────────┘    └────────────────────────┘    └────────────┘
     │                            │                   │
     │ 用户批准/拒绝              │                   │
     │───────────────────────────►│                   │
     │                             │                   │
     │                             │                   │
     │    ┌────────────────────┐ │                   │
     │    │ Event Broadcaster │◄─┘                   │
     │    │ (SSE)            │                      │
     │    └────────────────────┘                      │
     │                             │                  │
     │    ┌────────────────────┐ │                  │
     │    │ Native Messaging  │◄─┘                  │
     │    │ (Host)           │                     │
     │    └────────────────────┘                     │
     │                             │                  │
     └─────────────────────────────►│                  │
                                   │                  │
                                   │                  │
                                   │                  │
                                   │                  │
                                   │                  │
                                   │                  │
```

### 4.2 余额查询流程 (MCP直接交互)

```
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│AI Agent    │───►│ Native Host           │───►│区块链网络│
└────────────┘    └────────────────────────┘    └────────────┘
     │                   │                │
     │                   │                │
     │    ┌──────────────▼────────────┐   │
     │    │ MCP Server              │   │
     │    │ - HandleGetBalance()   │───┘
     │    └────────────────────────┘
     │                │
     │                │
     │    ┌────────────────────────┐
     │    │ Background Service     │
     │    │ Worker (Host)          │
     │    └────────────────────────┘
     │                │
     │                │
┌────────────┐    ┌────────────────────────┐
│ Browser UI │◄───│ 浏览器扩展 │
│(Popup)     │    │(Content Script)         │
└────────────┘    └────────────────────────┘
```

### 4.3 钱包导入流程 (通过浏览器扩展)

```
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│Browser Ext │──1─►│ 浏览器扩展 │──2─►│ Native Host│
│(Popup UI)  │    │(Content Script)         │    └────────────┘
└────────────┘    └────────────────────────┘          │
     │                            │                   │
     │ 用户输入助记词             │                   │
     │───────────────────────────►│                   │
     │                             │                   │
     │                             │                   │
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│ Browser UI │◄───│ Background Service     │◄───│Wallet Mgr  │
│(Popup)     │ SSE │ Worker (Host)          │    │             │
└────────────┘    └────────────────────────┘    └────────────┘
     │                             │                   │
     │ 用户确认                   │                   │
     │───────────────────────────►│                   │
     │                             │                   │
     │                             │                   │
     │    ┌────────────────────┐ │                   │
     │    │ Event Broadcaster │◄─┘                   │
     │    │ (SSE)            │                      │
     │    └────────────────────┘                      │
     │                             │                  │
     │    ┌────────────────────┐ │                  │
     │    │ Native Messaging  │◄─┘                  │
     │    │ (Host)           │                     │
     │    └────────────────────┘                     │
     │                             │                  │
     └─────────────────────────────►│                  │
                                   │                  │
                                   │                  │
                                   │                  │
                                   │                  │
                                   │                  │
                                   │                  │
```

## 5. 模块接口设计

### 5.1 Background与Content Script通信

```javascript
// Content Script
window.ethereum.request({
  method: 'eth_sendTransaction',
  params: [txParams],
});

// Background Service Worker
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  this.handleContentScriptMessage(request, sender, sendResponse);
  return true; // 保持消息通道开放
});
```

### 5.2 Background与Native Host通信

```javascript
// 发送消息到Native Host
this.port.postMessage({
  id: requestId,
  action: action,
  params: params,
});

// 接收Native Host消息
this.port.onMessage.addListener((response) => {
  this.handleNativeResponse(response);
});
```

### 5.3 Popup与Background通信

```javascript
// Popup UI发起请求
chrome.runtime.sendMessage({
    action: 'send_transaction',
    params: txParams
}, (response) => {
    if (response.error) {
        // 处理错误
    } else {
        // 显示交易状态
    }
});

// Background处理请求
handleContentScriptMessage(request, sender, sendResponse) {
    switch (request.action) {
        case 'request_accounts':
            // 处理账户请求
            break;

        case 'send_transaction':
            // 处理交易请求
            break;

        case 'personal_sign':
            // 处理签名请求
            break;
    }
}
```

### 5.4 Web3 Provider接口

```javascript
// Web3Provider实现
async request({ method, params }) {
    switch (method) {
        case 'eth_requestAccounts':
            return this.requestAccounts();

        case 'eth_accounts':
            return this.getAccounts();

        case 'eth_sendTransaction':
            return this.sendTransaction(params[0]);

        case 'personal_sign':
            return this.personalSign(params);

        default:
            throw new Error(`Method ${method} not supported`);
    }
}

// 交易请求
async sendTransaction(txParams) {
    // 发送到background script处理
    const response = await chrome.runtime.sendMessage({
        action: 'send_transaction',
        params: txParams
    });

    if (response.error) {
        throw new Error(response.error);
    }

    return response.hash;
}
```

## 6. 安全性设计

### 6.1 通信安全

- **Content Script与Background**:
  - 使用Chrome扩展消息传递API
  - 所有消息都经过Background验证
- **Background与Native Host**:
  - 使用Chrome Native Messaging API
  - 所有敏感操作都在Native Host完成

### 6.2 私钥安全

- 私钥永远不会离开Native Host
- 所有签名操作都在Native Host完成
- AI Agent只能通过MCP协议获得交易结果
- 用户通过Popup UI控制敏感操作

### 6.3 事件流

- **AI Agent**:
  - 通过MCP SSE流接收交易确认事件
  - 接收余额更新事件
- **Browser Extension**:
  - 通过Native Messaging接收交易确认
  - 通过Popup UI显示交易确认对话框
  - 通过SSE接收实时余额更新

## 7. 浏览器扩展与Native Host交互

### 7.1 交易签名流程

```
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│DEX Website │──1─►│ 浏览器扩展 │──2─►│ Native Host│
│(DApp)     │    │(Content Script)         │    └────────────┘
└────────────┘    └────────────────────────┘          │
     ▲                            │                   │
     │                            │                   │
     │                            │                   │
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│ Browser UI │──3─◄│ Background Service     │──4─◄│MCP Server  │
│(Popup)     │    │ Worker (Host)          │    │             │
└────────────┘    └────────────────────────┘    └────────────┘
     │                            │                   │
     │ 用户批准/拒绝             │                   │
     │───────────────────────────►│                   │
     │                             │                   │
     │                             │                   │
     │    ┌────────────────────┐  │                   │
     │    │ Event Broadcaster │◄─┘                   │
     │    │ (SSE)            │                      │
     │    └────────────────────┘                      │
     │                             │                  │
     │    ┌────────────────────┐  │                  │
     │    │ Native Messaging │◄─┘                  │
     │    │ (Host)          │                     │
     │    └────────────────────┘                     │
     │                             │                  │
     └─────────────────────────────►│                  │
                                   │                  │
                                   │                  │
                                   │                  │
                                   │                  │
                                   │                  │
                                   │                  │
```

### 7.2 事件订阅流程

```
┌────────────┐    ┌────────────────────────┐    ┌────────────┐
│AI Agent    │───►│ Native Host           │───►│MCP Server│
└────────────┘    └────────────────────────┘    └────────────┘
     │                   │                │
     │                   │                │
     │    ┌─────────────▼────────────┐   │
     │    │ Background Service       │   │
     │    │ Worker (Host)            │───┘
     │    └────────────────────────┘
     │                │
     │                │
┌────────────┐    ┌────────────────────────┐
│ Browser UI │◄───│ 浏览器扩展 │
│(Popup)     │    │(Content Script)         │
└────────────┘    └────────────────────────┘
```

## 8. 安全增强措施

1. **权限控制**
   - 所有敏感操作都需要用户确认
   - 钱包导入等操作通过Popup UI完成，不会暴露给AI Agent

2. **通信加密**
   - 使用Chrome Native Messaging提供的安全通道
   - 所有消息都经过扩展验证

3. **私钥隔离**
   - 私钥永远不会离开Native Host
   - AI Agent只能获得公钥和交易结果

4. **用户控制**
   - 所有交易都需要用户确认
   - 用户可以随时撤销AI Agent权限

5. **事件通知**
   - 通过SSE实时通知AI Agent
   - 通过Popup UI通知用户

6. **多链支持**
   - 动态切换区块链
   - 每个链使用独立的交易流程

7. **交易确认**
   - 所有交易都需要AI Agent和用户双重确认
   - 交易细节对用户透明

## 9. 优势分析

### 9.1 安全优势

- **私钥隔离**: 私钥始终在Native Host内部
- **用户控制**: 所有敏感操作都通过Popup UI确认
- **AI Agent限制**: 不允许执行钱包导入等敏感操作
- **通信安全**: 使用Chrome Native Messaging和SSE加密事件流

### 9.2 架构优势

- **模块化**: 清晰的模块划分
- **兼容性**: 支持现有DApp
- **可扩展性**: 易于添加新功能
- **实时性**: 通过SSE实现快速事件通知

### 9.3 用户体验优势

- **无缝集成**: DApp无需特殊改造
- **智能交易**: AI Agent可以预判交易
- **安全确认**: 用户可以审核AI Agent建议
- **实时监控**: 钱包状态实时更新
- **直观界面**: 提供清晰的交易确认对话框

## 10. 开发计划

### Phase 1: 基础架构

- [x] manifest.json v3配置
- [x] Background Service Worker实现
- [x] Content Script实现
- [x] Web3提供者注入框架
- [x] Popup界面基本结构
- [x] Native Messaging集成

### Phase 2: 功能实现

- [ ] 完整的Web3提供者实现
- [ ] 交易确认流程
- [ ] 余额查询和更新
- [ ] 事件通知系统
- [ ] 钱包管理界面
- [ ] 链切换支持

### Phase 3: 安全增强

- [ ] 多重签名支持
- [ ] 硬件钱包集成
- [ ] 风控系统
- [ ] 权限管理界面

### Phase 4: 高级功能

- [ ] 智能交易策略
- [ ] 交易历史查看
- [ ] 多链支持
- [ ] DeFi协议集成

这个设计文档提供了完整的浏览器扩展架构，确保了与Native Host的安全通信，同时保持了良好的用户体验和扩展性。
