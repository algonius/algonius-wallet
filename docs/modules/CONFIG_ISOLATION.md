# Configuration Isolation for Integration Tests

This document describes the configuration isolation system implemented to allow concurrent execution of integration tests without conflicts.

## Problem Statement

Previously, integration tests had the following issues:

1. **Shared Configuration**: All tests used the same `~/.algonius-wallet/` directory
2. **Process Conflicts**: Single instance check prevented concurrent test execution  
3. **Data Corruption**: Tests could interfere with each other's wallet data
4. **RUN_MODE Dependency**: Used `RUN_MODE=test` which was not flexible enough

## Solution Overview

The new system uses environment variables to provide complete isolation:

### Environment Variables

#### `ALGONIUS_WALLET_HOME`
- **Purpose**: Override the default wallet home directory (`~/.algonius-wallet/`)
- **Usage**: Each test gets its own isolated directory
- **Effect**: Separates configuration, wallet data, and logs

#### Benefits of New Approach
- ✅ **Clean Separation**: No RUN_MODE dependency
- ✅ **Concurrent Testing**: Multiple tests can run simultaneously
- ✅ **Isolation**: Each test has its own data directory
- ✅ **Flexibility**: Can be used for different environments (dev, staging, etc.)

## Implementation Details

### 1. Configuration System (`pkg/config/config.go`)

```go
// getWalletHomeDir returns the wallet home directory, respecting environment override
func getWalletHomeDir() string {
    // Check environment variable first
    if homeDir := os.Getenv("ALGONIUS_WALLET_HOME"); homeDir != "" {
        return homeDir
    }
    
    // Default path
    userHome, err := os.UserHomeDir()
    if err != nil {
        return "."
    }
    return filepath.Join(userHome, ".algonius-wallet")
}
```

**Changes made:**
- Removed `RUN_MODE=test` logic from `LoadConfigWithFallback()`
- Modified `DefaultConfig()` to use `getWalletHomeDir()`
- Updated `GetConfigPath()` to respect `ALGONIUS_WALLET_HOME`

### 2. Wallet Manager (`pkg/wallet/manager.go`)

```go
func NewWalletManager() *WalletManager {
    // Get wallet directory from environment or default
    walletHomeDir := getWalletHomeDir()
    walletDir := filepath.Join(walletHomeDir, "wallets")
    
    // Create wallet directory if it doesn't exist
    os.MkdirAll(walletDir, 0700)
    
    return &WalletManager{
        walletDir: walletDir,
        // ... other fields
    }
}
```

**Changes made:**
- Wallet manager now respects `ALGONIUS_WALLET_HOME`
- Automatic creation of isolated wallet directories

### 3. Process Management (`cmd/main.go`)

```go
// Skip process locking if ALGONIUS_WALLET_HOME is set (indicating test/isolated environment)
isIsolatedEnvironment := os.Getenv("ALGONIUS_WALLET_HOME") != ""

if !isIsolatedEnvironment {
    // Try to acquire PID file lock to prevent multiple instances
    locked, err := process.LockPIDFile()
    // ...
}
```

**Changes made:**
- Replaced `RUN_MODE=test` check with `ALGONIUS_WALLET_HOME` detection
- Allows multiple isolated instances to run concurrently
- Maintains single-instance protection for production use

### 4. Test Environment (`tests/integration/env/mcp_host_test_environment.go`)

```go
func (env *McpHostTestEnvironment) startMcpHost(ctx context.Context) error {
    // Use test data directory as isolated wallet home
    testWalletHome := env.testDataDir
    
    // Prepare environment variables with isolated paths
    environ := append(os.Environ(),
        fmt.Sprintf("ALGONIUS_WALLET_HOME=%s", testWalletHome),
        fmt.Sprintf("LOG_FILE=%s", env.logFilePath),
        "LOG_LEVEL=debug",
    )
    
    // Start the MCP host process
    env.hostProcess = exec.CommandContext(ctx, "../../bin/algonius-wallet-host")
    env.hostProcess.Env = environ
    // ...
}
```

**Changes made:**
- Each test gets a unique temporary directory
- Environment variables provide complete isolation
- Removed dependency on `RUN_MODE=test`

## Directory Structure

### Production
```
~/.algonius-wallet/
├── config.yaml
├── wallets/
│   └── wallet.json
└── logs/
    └── wallet.log
```

### Test (Isolated)
```
/tmp/mcp-host-test/test-1738123456/
├── config.yaml         # Test-specific config
├── wallets/             # Isolated wallet data
│   └── wallet.json
└── logs/                # Test-specific logs
    └── mcp-host.log
```

## Benefits Achieved

### ✅ Concurrent Test Execution
Tests can now run in parallel without conflicts:

```bash
# These can run simultaneously
go test -v ./tests/integration -run TestUnlockWalletHandler_Success &
go test -v ./tests/integration -run TestUnlockWalletHandler_WrongPassword &
```

### ✅ Complete Isolation
Each test has its own:
- Configuration file
- Wallet storage directory  
- Log files
- Process instance

### ✅ No RUN_MODE Dependency
The system is more flexible and doesn't rely on global modes:

```bash
# Production use (default behavior)
./algonius-wallet-host

# Isolated environment (custom home)
ALGONIUS_WALLET_HOME=/custom/path ./algonius-wallet-host

# Test use (automatic isolation)
# Set by test framework automatically
```

### ✅ Backward Compatibility
- Production behavior unchanged
- Existing configurations continue to work
- Default paths remain the same

## Usage Examples

### Manual Testing
```bash
# Create isolated test environment
export ALGONIUS_WALLET_HOME=/tmp/test-wallet
./algonius-wallet-host

# Data will be stored in:
# /tmp/test-wallet/config.yaml
# /tmp/test-wallet/wallets/wallet.json
```

### Integration Tests
```bash
# Run single test
go test -v ./tests/integration -run TestUnlockWalletHandler_Success

# Run all unlock tests concurrently  
go test -v ./tests/integration -run "TestUnlockWalletHandler.*" -parallel 4
```

### Development Environment
```bash
# Use separate dev environment
export ALGONIUS_WALLET_HOME=~/.algonius-wallet-dev
./algonius-wallet-host
```

## Migration Notes

### For Developers
- Tests now run faster due to elimination of artificial delays
- Multiple tests can run concurrently
- Each test starts with a clean environment

### For CI/CD
- Parallel test execution is now safe
- No need for test serialization
- Improved test reliability

### For Production
- No changes required
- Default behavior is preserved
- Configuration paths remain the same

## Future Enhancements

### Potential Improvements
1. **Environment Templates**: Pre-configured test environments
2. **Cleanup Automation**: Automatic removal of old test directories
3. **Configuration Validation**: Ensure test configs are properly isolated
4. **Performance Monitoring**: Track test execution times with isolation

### Configuration Options
Consider adding more environment variables for fine-grained control:
- `ALGONIUS_WALLET_LOG_LEVEL`: Override log level
- `ALGONIUS_WALLET_NETWORK`: Override network mode
- `ALGONIUS_WALLET_TIMEOUT`: Configure operation timeouts

This isolation system provides a robust foundation for reliable, concurrent integration testing while maintaining production stability.