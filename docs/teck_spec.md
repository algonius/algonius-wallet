# Algonius Wallet 技术规格概要

## 1. 项目架构总览

Algonius Wallet 由以下核心组件组成：

- **AI Agent**  
  通过 MCP 协议与 Native Host 交互，实现自动化交易与智能决策。

- **Browser Extension**  
  负责 Web3 注入、DApp 交互、用户界面与 Native Host 的安全通信。  
  详细设计见 [docs/modules/browser_extension_design.md](./modules/browser_extension_design.md)

- **Native Host**  
  负责多链钱包管理、交易签名、MCP Server 实现、与浏览器扩展的 Native Messaging 通信。  
  详细设计见 [docs/modules/native_host_design.md](./modules/native_host_design.md)

## 2. 组件关系与通信

```
┌──────────┐      ┌──────────────┐      ┌────────────┐
│ AI Agent │◄───►│ Native Host  │◄───►│ Browser Ext│
└──────────┘      └──────────────┘      └────────────┘
      ▲                ▲                      │
      │                │                      │
      └─────MCP────────┘           Web3/DApp  │
                                 注入/交互     │
```

- **AI Agent** 通过 MCP 工具/资源与 Native Host 交互，不能直接访问私钥或敏感操作。
- **Browser Extension** 通过 Native Messaging 与 Native Host 通信，负责前端交互与安全边界。
- **Native Host** 作为安全核心，负责所有链上操作、密钥管理和事件广播。

## 2.1 子系统职责与接口摘要

### AI Agent

- **主要职责**：
  - 通过 MCP 协议与 Native Host 交互，实现自动化交易、余额查询、事件订阅等智能决策功能。
  - 不直接访问私钥或敏感操作，仅能调用受控的工具/资源接口。
- **关键接口摘要**（MCP 工具/资源）：
  - `get_balance`：查询指定地址/链的余额
  - `send_transaction`：发起链上转账
  - `sign_message`：消息签名（受限）
  - `swap_tokens`：代币兑换
  - `get_transactions`：查询交易历史
  - `confirm_transaction`：AI 决策确认
  - 事件订阅（SSE）：交易状态、余额变更等

### Browser Extension

- **主要职责**：
  - 注入 Web3 Provider，兼容 DApp 生态，拦截并转发 Web3 请求
  - 管理用户界面（Popup UI），处理钱包管理、交易确认、设置等
  - 通过 Native Messaging 与 Native Host 安全通信，转发敏感操作
- **关键接口摘要**：
  - Content Script ⇄ Background Service Worker：Web3 请求转发（如 `eth_sendTransaction`、`personal_sign`）
  - Popup UI ⇄ Background：钱包状态、余额、交易确认、设置等消息
  - Background ⇄ Native Host：交易签名、账户管理、事件订阅等消息
  - Web3 Provider（EIP-1193）：`requestAccounts`、`sendTransaction`、`personalSign`、`switchChain` 等

### Native Host

- **主要职责**：
  - 作为安全核心，管理多链钱包、私钥、交易签名、链上交互
  - 实现 MCP Server，提供工具/资源接口与事件流
  - 通过 Native Messaging 与浏览器扩展通信，处理敏感操作
  - 事件广播，推送链上状态变更
- **关键接口摘要**：
  - MCP Server 工具/资源接口（见上文 AI Agent 部分）
  - Native Messaging 接口：`import_wallet`、`export_wallet`、`backup_wallet`、`sign_transaction`、`get_accounts`、`get_balance`、`event_subscription` 等
  - Wallet Manager 接口：`NewWallet`、`ImportWallet`、`SendTransaction`、`GetBalance`、`SignTransaction` 等
  - Event Broadcaster：SSE 事件推送

## 3. 主要业务流程

- **DApp 交易确认**  
  DApp → Content Script → Background → Native Host → AI Agent 决策 → Native Host 签名 → DApp

- **余额/状态查询**  
  AI Agent/前端 → Native Host → 区块链 → 结果回传

- **钱包导入/管理**  
  仅允许通过 Browser Extension UI 进行，Native Host 本地加密存储，MCP 不暴露敏感导入接口。

## 4. 组件详细设计引用

- **Browser Extension 详细设计**  
  见 [docs/modules/browser_extension_design.md](./modules/browser_extension_design.md)

- **Native Host 详细设计**  
  见 [docs/modules/native_host_design.md](./modules/native_host_design.md)

---

该概要文档只描述组件级架构、关系和主流程，所有实现细节和接口规范请参见 docs/modules/ 下的详细设计文档。

## 测试方案引用

- **E2E 测试方案**  
  见 [docs/tests/e2e_test_plan.md](./tests/e2e_test_plan.md)
