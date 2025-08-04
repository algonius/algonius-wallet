# Algonius Wallet E2E 测试方案

## 1. 测试目标

- 验证 Algonius Wallet 全链路功能的正确性、健壮性与安全性
- 覆盖 DApp 交互、Browser Extension、Native Host、AI Agent、区块链模拟等关键路径
- 保证用户体验、数据一致性与安全边界

## 2. 测试环境准备

- **浏览器**：Chrome（推荐最新版，支持 Manifest V3）
- **区块链模拟器**：Ganache/Hardhat/Ethereum Testnet
- **Native Host**：本地编译并运行
- **AI Agent**：本地或模拟服务
- **测试 DApp**：自建或选用开源 DEX/DApp
- **自动化工具**：Playwright/Puppeteer + Mocha/Jest + Go test
- **测试账户/助记词**：预置测试钱包

## 3. 测试工具与依赖

- Node.js、npm/yarn
- Go 环境
- ChromeDriver/Playwright
- Ganache/Hardhat
- 相关依赖包

## 4. 核心 E2E 测试用例

### 4.1 DApp 交易确认流程

- DApp 发起交易请求（如 swap/转账）
- Content Script 拦截并转发
- Background Service Worker 路由到 Native Host
- AI Agent 决策（可模拟自动/手动确认）
- Popup UI 弹窗，用户确认/拒绝
- Native Host 完成签名并广播交易
- DApp 获取交易结果
- 验证链上交易状态

### 4.2 余额/状态查询

- DApp/AI Agent/Popup 查询余额
- Native Host 查询链上余额
- 返回并展示余额
- 验证数据一致性

### 4.3 钱包导入/切换

- Popup UI 导入新钱包（助记词/私钥）
- Native Host 本地加密存储
- 切换账户，验证地址与余额
- 验证敏感操作无越权

### 4.4 事件推送与实时更新

- 触发链上事件（如转账、余额变更）
- Native Host 通过 SSE/Native Messaging 推送事件
- Popup UI/AI Agent 实时接收并展示
- 验证事件流畅通与准确性

### 4.5 异常与安全场景

- 非法交易请求拦截
- 助记词格式错误处理
- 用户拒绝交易
- AI Agent 权限撤销
- 网络异常/Native Host 崩溃恢复

## 5. 自动化与回归建议

- 推荐使用 Playwright/Puppeteer 脚本自动化 DApp 操作与 UI 交互
- Go test 脚本自动化 Native Host 接口测试
- Mock AI Agent 行为，覆盖自动/手动决策分支
- 每次主版本发布前全量回归

## 6. 数据准备与清理

- 预置测试账户、助记词、链上资产
- 测试结束后清理本地钱包与链上测试数据

## 7. 预期结果与通过标准

- 所有功能路径均能正确完成，数据一致
- 所有安全边界无越权、无敏感数据泄露
- 事件推送实时、准确
- UI 交互流畅，异常场景有明确提示

## 8. 扩展性与维护建议

- 用例可扩展至多链、DeFi、硬件钱包等新功能
- 测试脚本与用例文档化，便于团队协作与持续集成

---

> **说明**：如需详细用例模板或自动化脚本示例，可在本目录下补充 `e2e_cases.md`、`automation_sample.js` 等文件。
