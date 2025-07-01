# native-host 测试策略说明

本说明文件明确 native-host 的单元测试与集成测试组织方式、目录结构与最佳实践。

---

## 1. 单元测试（Unit Test）

- **位置**：与源码同目录（即各 pkg/_ 子包下，采用 _\_test.go 文件）
- **原则**：
  - 每个核心模块（如 mcp、wallet、messaging、api、event、config、security、utils）都应有对应的单元测试文件。
  - 重点覆盖纯函数、数据结构、接口实现、边界条件、错误处理等。
  - Mock 外部依赖，聚焦单一功能的正确性和健壮性。
  - **断言库统一使用 [github.com/stretchr/testify/assert](https://github.com/stretchr/testify)**
    - 推荐在每个 \*\_test.go 文件头部引入：
      ```go
      import (
          "testing"
          "github.com/stretchr/testify/assert"
      )
      ```
    - 常用断言示例：
      ```go
      assert.Equal(t, expected, actual)
      assert.NoError(t, err)
      ```
  - **单元测试代码覆盖率要求 90% 及以上**
    - 所有核心逻辑需有充分测试，建议在本地和 CI 中强制校验覆盖率。
- **示例结构**：
  ```
  native/pkg/wallet/manager.go
  native/pkg/wallet/manager_test.go
  native/pkg/mcp/server.go
  native/pkg/mcp/server_test.go
  ...
  ```

---

## 2. 集成测试（Integration Test）

- **位置**：统一放置于 native/tests/integration 目录下
- **原则**：
  - 跨模块/跨进程的功能链路测试，如 MCP 工具端到端调用、Native Messaging 与浏览器扩展交互、事件推送全链路等。
  - 可用 Go 的 testing 包，也可结合 shell 脚本、模拟外部进程（如 mock browser extension/AI Agent）。
  - 集成测试用例命名清晰，便于持续集成自动发现和执行。
- **示例结构**：
  ```
  native/tests/integration/mcp_integration_test.go
  native/tests/integration/native_messaging_test.go
  native/tests/integration/event_broadcast_test.go
  ...
  ```

---

### 2.1 设计原则与架构

- **进程级真实环境**：测试用例以真实 Go 进程方式启动 native-host，模拟生产环境。
- **双协议链路**：同时覆盖 Native Messaging（Chrome Extension 通道）与 HTTP/SSE（MCP 客户端通道）。
- **端到端流**：测试用例应覆盖从“模拟扩展”到“host”再到“模拟 MCP 客户端”的全链路。
- **资源与工具注册**：验证所有资源、工具注册与协议一致性。
- **健壮性与错误处理**：覆盖异常输入、协议违规、并发等边界场景。

#### 架构图

```
┌─────────────────┐      ┌────────────────┐      ┌────────────────────┐
│  Mock Chrome    │      │   MCP Host     │      │    Mock MCP        │
│  Extension      │ ──── │   Process      │ ──── │    Client          │
│  (Go)           │      │   (Under Test) │      │    (HTTP/SSE)      │
└─────────────────┘      └────────────────┘      └────────────────────┘
     Native Messaging         SSE/HTTP Protocol
     (stdin/stdout)
```

---

### 2.2 关键测试场景与代码示例

- 进程生命周期与健康检查
- 资源注册与读取
- 工具注册与调用
- 协议合规性与错误处理
- 并发与性能
- Native Messaging 与 SSE 事件流全链路

#### 典型测试环境与用例代码（节选）

<details>
<summary>测试环境管理与进程启动</summary>

```go
type McpHostTestEnvironment struct {
    hostProcess   *exec.Cmd
    mcpClient     *MockMcpSSEClient
    nativeMsg     *NativeMessagingManager
    port          int
    baseURL       string
    basePath      string
    logFilePath   string
    testDataDir   string
}

// ... 详见项目文档
```

</details>

<details>
<summary>端到端流程、资源、工具、错误处理等测试用例</summary>

```go
func TestProcessLifecycle(t *testing.T) { /* ... */ }
func TestBrowserStateResource(t *testing.T) { /* ... */ }
func TestNavigateToTool(t *testing.T) { /* ... */ }
func TestErrorHandling(t *testing.T) { /* ... */ }
```

</details>

<details>
<summary>Mock Chrome Extension 与 Mock MCP Client 设计</summary>

```go
type NativeMessagingManager struct { /* ... */ }
type MockMcpSSEClient struct { /* ... */ }
```

</details>

---

### 2.3 测试组织与 CI 集成建议

- **测试用例组织**：按功能分文件，如 lifecycle、resources、tools、protocol、errors、performance、native_messaging。
- **Makefile/CI 集成**：提供 test-integration、test-integration-coverage 等命令，便于一键运行和覆盖率统计。

#### Makefile 示例

```makefile
.PHONY: test-integration test-integration-verbose test-integration-clean

test-integration: build
	@echo "Running integration tests..."
	cd tests/integration && go test -v -timeout=5m ./...

test-integration-coverage: build
	@echo "Running integration tests with coverage..."
	cd tests/integration && go test -v -timeout=5m -coverprofile=coverage.out ./...
	cd tests/integration && go tool cover -html=coverage.out -o coverage.html
```

---

如需完整代码模板、辅助工具实现、详细用例可参考 native/tests/integration 目录及相关文档。

---

## 3. 组织与运行建议

- **单元测试**：`go test ./...` 可自动递归运行所有 \*\_test.go
- **集成测试**：可单独运行 `go test ./tests/integration/...`，或在 CI/CD 脚本中分阶段执行
- **Mock/Stub**：推荐使用 testify/mock、Go 标准库 testing 等工具
- **覆盖率**：
  - 单元测试覆盖率要求 90% 及以上，集成测试建议生成覆盖率报告，便于质量跟踪
  - 推荐命令：
    ```bash
    go test ./pkg/... -coverprofile=coverage.out
    go tool cover -func=coverage.out
    ```
  - 可在 CI/CD 脚本中设置覆盖率阈值，未达标时构建失败

---

如需具体测试模板或用例设计方案，可进一步细化每个模块的测试目标与示例代码。
