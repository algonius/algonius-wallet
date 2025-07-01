# Browser Extension Native Messaging API

本文档定义了 Native Host 为浏览器扩展提供的 Native Messaging RPC 接口。

## 概述

Native Host 通过 Chrome Native Messaging 协议为浏览器扩展提供安全的钱包操作接口。这些接口只能通过 Native Messaging 访问，不暴露给 AI Agents。

## 通信协议

### 消息格式

```json
{
  "type": "rpc_request",
  "id": "string (unique request ID)",
  "method": "string (RPC method name)",
  "params": "object (method parameters)"
}
```

### 响应格式

```json
{
  "type": "rpc_response",
  "id": "string (matching request ID)",
  "result": "object (method result)",
  "error": {
    "code": "number",
    "message": "string"
  }
}
```

## 接口列表

### 1. import_wallet

导入钱包使用助记词。

**参数:**

```json
{
  "mnemonic": "string (required)",
  "password": "string (required)",
  "chain": "string (required, enum: [\"ethereum\", \"bsc\"])",
  "derivation_path": "string (optional, default: \"m/44'/60'/0'/0/0\")"
}
```

**返回:**

```json
{
  "address": "string",
  "public_key": "string",
  "imported_at": "number (timestamp)"
}
```

**错误码:**

- `-32001`: 无效助记词
- `-32002`: 密码强度不足
- `-32003`: 不支持的链
- `-32004`: 钱包已存在
- `-32005`: 存储加密失败

**相关 Issue:** [#008](../issues/008-implement-import-wallet-native-messaging.md)

---

### 2. export_wallet

导出钱包私钥或助记词。

**参数:**

```json
{
  "address": "string (required)",
  "password": "string (required)",
  "export_type": "string (required, enum: [\"private_key\", \"mnemonic\"])",
  "encryption_key": "string (optional)"
}
```

**返回:**

```json
{
  "export_data": "string (encrypted)",
  "export_type": "string",
  "exported_at": "number (timestamp)",
  "expires_at": "number (timestamp)"
}
```

**错误码:**

- `-32011`: 无效密码
- `-32012`: 钱包未找到
- `-32013`: 不支持的导出类型
- `-32014`: 导出尝试次数过多
- `-32015`: 导出加密失败

**相关 Issue:** [#009](../issues/009-implement-export-wallet-native-messaging.md)

---

### 3. get_wallet_info

获取钱包信息包括余额和状态。

**参数:**

```json
{
  "chain": "string (optional, enum: [\"ethereum\", \"bsc\", \"all\"])",
  "address": "string (optional)",
  "include_balances": "boolean (optional, default: true)",
  "include_tokens": "boolean (optional, default: false)"
}
```

**返回:**

```json
{
  "wallets": [
    {
      "address": "string",
      "chain": "string",
      "balance": "string",
      "tokens": [
        {
          "contract": "string",
          "symbol": "string",
          "balance": "string",
          "decimals": "number"
        }
      ],
      "created_at": "number (timestamp)",
      "alias": "string",
      "is_active": "boolean"
    }
  ],
  "total_count": "number",
  "last_updated": "number (timestamp)"
}
```

**错误码:**

- `-32021`: 无效链参数
- `-32022`: 无效地址格式
- `-32023`: 钱包未找到
- `-32024`: 余额获取失败

**相关 Issue:** [#010](../issues/010-implement-get-wallet-info-native-messaging.md)

---

### 4. send_transaction

发送区块链交易（需要用户确认）。

**参数:**

```json
{
  "from": "string (required)",
  "to": "string (required)",
  "amount": "string (required)",
  "chain": "string (required, enum: [\"ethereum\", \"bsc\"])",
  "token": "string (optional, contract address)",
  "gas_limit": "number (optional)",
  "gas_price": "string (optional)",
  "data": "string (optional, hex data)",
  "password": "string (required)"
}
```

**返回:**

```json
{
  "transaction_hash": "string",
  "status": "string",
  "gas_used": "number",
  "block_number": "number (optional)",
  "sent_at": "number (timestamp)"
}
```

**错误码:**

- `-32031`: 无效密码
- `-32032`: 余额不足
- `-32033`: 无效接收地址
- `-32034`: 用户拒绝交易
- `-32035`: Gas 估算失败
- `-32036`: 交易广播失败

**相关 Issue:** [#011](../issues/011-implement-send-transaction-native-messaging.md)

---

## 安全考虑

### 身份验证

- 所有敏感操作需要密码验证
- 实施操作频率限制
- 记录所有敏感操作的审计日志

### 数据保护

- 私钥和助记词使用强加密存储
- 从不在日志中记录私密信息
- 导出数据临时加密

### 用户确认

- 所有交易需要用户明确确认
- 显示完整交易详情供用户审核
- 实施交易金额限制

## 实现状态

| 接口             | 状态   | 优先级 | Issue |
| ---------------- | ------ | ------ | ----- |
| import_wallet    | 待实现 | 高     | #008  |
| export_wallet    | 待实现 | 高     | #009  |
| get_wallet_info  | 待实现 | 中     | #010  |
| send_transaction | 待实现 | 高     | #011  |

## 相关文档

- [技术规格](../teck_spec.md)
- [Native Host MCP API](./native_host_mcp_api.md)
- [Native Messaging 实现](../../native/pkg/messaging/native.go)
