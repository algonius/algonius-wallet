# Native Host MCP 工具与资源接口一览（开发实现用）

本文件列出 native_host 需要暴露给 AI Agent 的所有工具（Tools）和资源（Resources）接口，均为实际开发所需，格式完全对齐 MCP 规范，便于直接注册与实现。

---

## 1. 工具（Tools）

每个工具均需注册如下 schema，所有字段均为必需：

- `name`：工具唯一标识
- `description`：用途说明
- `input_schema`：输入参数结构（JSON Schema）
- `output_schema`：输出结果结构（JSON Schema）
- `error_schema`：错误结构（JSON Schema）
- `security`：安全属性（如“需用户授权”）

### 1.1 create_wallet

```json
{
  "name": "create_wallet",
  "description": "创建新钱包（本地生成私钥）",
  "input_schema": {
    "type": "object",
    "properties": {
      "chain": { "type": "string", "description": "链标识，如 ETH" }
    },
    "required": ["chain"]
  },
  "output_schema": {
    "type": "object",
    "properties": {
      "address": { "type": "string" },
      "public_key": { "type": "string" }
    },
    "required": ["address", "public_key"]
  },
  "error_schema": {
    "type": "object",
    "properties": {
      "code": { "type": "integer" },
      "message": { "type": "string" }
    },
    "required": ["code", "message"]
  },
  "security": "需用户授权"
}
```

### 1.2 get_balance

```json
{
  "name": "get_balance",
  "description": "查询钱包余额",
  "input_schema": {
    "type": "object",
    "properties": {
      "address": { "type": "string" },
      "token": { "type": "string", "description": "主币或代币合约地址" }
    },
    "required": ["address", "token"]
  },
  "output_schema": {
    "type": "object",
    "properties": {
      "balance": { "type": "string" }
    },
    "required": ["balance"]
  },
  "error_schema": {
    "type": "object",
    "properties": {
      "code": { "type": "integer" },
      "message": { "type": "string" }
    },
    "required": ["code", "message"]
  },
  "security": "需用户授权"
}
```

### 1.3 send_transaction

```json
{
  "name": "send_transaction",
  "description": "发起链上转账",
  "input_schema": {
    "type": "object",
    "properties": {
      "from": { "type": "string" },
      "to": { "type": "string" },
      "value": { "type": "string" },
      "token": { "type": "string" },
      "gas_limit": { "type": "integer" },
      "gas_price": { "type": "string" },
      "data": { "type": "string", "nullable": true }
    },
    "required": ["from", "to", "value", "token"]
  },
  "output_schema": {
    "type": "object",
    "properties": {
      "tx_hash": { "type": "string" },
      "status": { "type": "string", "enum": ["pending", "confirmed", "failed"] }
    },
    "required": ["tx_hash", "status"]
  },
  "error_schema": {
    "type": "object",
    "properties": {
      "code": { "type": "integer" },
      "message": { "type": "string" }
    },
    "required": ["code", "message"]
  },
  "security": "需用户授权"
}
```

### 1.4 confirm_transaction

```json
{
  "name": "confirm_transaction",
  "description": "查询交易状态",
  "input_schema": {
    "type": "object",
    "properties": {
      "tx_hash": { "type": "string" }
    },
    "required": ["tx_hash"]
  },
  "output_schema": {
    "type": "object",
    "properties": {
      "status": { "type": "string", "enum": ["pending", "confirmed", "failed"] }
    },
    "required": ["status"]
  },
  "error_schema": {
    "type": "object",
    "properties": {
      "code": { "type": "integer" },
      "message": { "type": "string" }
    },
    "required": ["code", "message"]
  },
  "security": "无需授权"
}
```

### 1.5 get_transactions

```json
{
  "name": "get_transactions",
  "description": "查询钱包历史交易",
  "input_schema": {
    "type": "object",
    "properties": {
      "address": { "type": "string" },
      "limit": { "type": "integer", "default": 10 }
    },
    "required": ["address"]
  },
  "output_schema": {
    "type": "object",
    "properties": {
      "transactions": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "hash": { "type": "string" },
            "from": { "type": "string" },
            "to": { "type": "string" },
            "value": { "type": "string" },
            "status": { "type": "string" }
          },
          "required": ["hash", "from", "to", "value", "status"]
        }
      }
    },
    "required": ["transactions"]
  },
  "error_schema": {
    "type": "object",
    "properties": {
      "code": { "type": "integer" },
      "message": { "type": "string" }
    },
    "required": ["code", "message"]
  },
  "security": "需用户授权"
}
```

### 1.6 sign_message

```json
{
  "name": "sign_message",
  "description": "对消息进行钱包签名",
  "input_schema": {
    "type": "object",
    "properties": {
      "address": { "type": "string" },
      "message": { "type": "string" }
    },
    "required": ["address", "message"]
  },
  "output_schema": {
    "type": "object",
    "properties": {
      "signature": { "type": "string" }
    },
    "required": ["signature"]
  },
  "error_schema": {
    "type": "object",
    "properties": {
      "code": { "type": "integer" },
      "message": { "type": "string" }
    },
    "required": ["code", "message"]
  },
  "security": "需用户授权"
}
```

### 1.7 swap_tokens

```json
{
  "name": "swap_tokens",
  "description": "代币兑换（跨链/同链）",
  "input_schema": {
    "type": "object",
    "properties": {
      "from_address": { "type": "string" },
      "to_token": { "type": "string" },
      "amount": { "type": "string" }
    },
    "required": ["from_address", "to_token", "amount"]
  },
  "output_schema": {
    "type": "object",
    "properties": {
      "swap_tx_hash": { "type": "string" },
      "status": { "type": "string" }
    },
    "required": ["swap_tx_hash", "status"]
  },
  "error_schema": {
    "type": "object",
    "properties": {
      "code": { "type": "integer" },
      "message": { "type": "string" }
    },
    "required": ["code", "message"]
  },
  "security": "需用户授权"
}
```

---

## 2. 资源（Resources）

每个资源均需注册如下 schema：

- `name`：资源唯一标识
- `description`：用途说明
- `schema`：资源数据结构（JSON Schema）

### 2.1 wallet_status

```json
{
  "name": "wallet_status",
  "description": "查询钱包状态",
  "schema": {
    "type": "object",
    "properties": {
      "address": { "type": "string" },
      "public_key": { "type": "string" },
      "ready": { "type": "boolean" },
      "chains": { "type": "object", "additionalProperties": { "type": "boolean" } },
      "last_used": { "type": "integer" }
    },
    "required": ["address", "public_key", "ready"]
  }
}
```

### 2.2 supported_chains

```json
{
  "name": "supported_chains",
  "description": "支持的区块链网络列表",
  "schema": {
    "type": "object",
    "properties": {
      "chains": {
        "type": "array",
        "items": { "type": "string" }
      }
    },
    "required": ["chains"]
  }
}
```

### 2.3 supported_tokens

```json
{
  "name": "supported_tokens",
  "description": "指定链支持的代币列表",
  "schema": {
    "type": "object",
    "properties": {
      "chain": { "type": "string" },
      "tokens": {
        "type": "array",
        "items": { "type": "string" }
      }
    },
    "required": ["chain", "tokens"]
  }
}
```

---

**说明：**

- 所有工具/资源接口均需严格注册，参数、返回值、错误结构与 MCP 规范一致。
- 工具接口需声明安全属性，敏感操作需用户授权。
- 资源接口支持只读快照和订阅推送，推送需用户同意。
- 禁止暴露高危接口（如导出私钥、导入助记词等）。
