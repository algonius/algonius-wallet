package integration

import (
	"context"
	"testing"
	"time"

	"github.com/algonius/algonius-wallet/native/tests/integration/env"
	"github.com/stretchr/testify/require"
)

func TestUnlockWalletHandler_Success(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	nativeMsg := testEnv.GetNativeMsg()
	require.NotNil(t, nativeMsg, "Native messaging manager should not be nil")

	// First, import a wallet to have something to unlock
	importParams := map[string]interface{}{
		"mnemonic":        "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		"password":        "TestPassword123!",
		"chain":           "ethereum",
		"derivationPath":  "m/44'/60'/0'/0/0",
	}

	importResponse, err := nativeMsg.RpcRequest(ctx, "import_wallet", importParams)
	require.NoError(t, err, "failed to import wallet")
	require.NotNil(t, importResponse, "import wallet response should not be nil")
	
	// Verify import was successful
	require.NotContains(t, importResponse, "error", "import wallet should not return error")
	require.Contains(t, importResponse, "result", "import wallet should return result")

	result, ok := importResponse["result"].(map[string]interface{})
	require.True(t, ok, "result should be a map")
	
	address, ok := result["address"].(string)
	require.True(t, ok, "address should be a string")
	require.NotEmpty(t, address, "address should not be empty")
	require.True(t, len(address) > 10, "address should be a valid length")

	// Wait a moment for wallet to be saved
	time.Sleep(100 * time.Millisecond)

	// Lock the wallet first
	lockResponse, err := nativeMsg.RpcRequest(ctx, "lock_wallet", map[string]interface{}{})
	require.NoError(t, err, "failed to lock wallet")
	require.NotContains(t, lockResponse, "error", "lock wallet should not return error")

	// Verify wallet is locked
	statusResponse, err := nativeMsg.RpcRequest(ctx, "wallet_status", map[string]interface{}{})
	require.NoError(t, err, "failed to get wallet status")
	require.NotContains(t, statusResponse, "error", "wallet status should not return error")
	
	statusResult, ok := statusResponse["result"].(map[string]interface{})
	require.True(t, ok, "status result should be a map")
	require.True(t, statusResult["has_wallet"].(bool), "should have wallet")
	require.False(t, statusResult["is_unlocked"].(bool), "wallet should be locked")

	// Now test unlocking with correct password
	unlockParams := map[string]interface{}{
		"password": "TestPassword123!",
	}

	unlockResponse, err := nativeMsg.RpcRequest(ctx, "unlock_wallet", unlockParams)
	require.NoError(t, err, "failed to unlock wallet")
	require.NotNil(t, unlockResponse, "unlock wallet response should not be nil")
	require.NotContains(t, unlockResponse, "error", "unlock wallet should not return error")
	require.Contains(t, unlockResponse, "result", "unlock wallet should return result")

	unlockResult, ok := unlockResponse["result"].(map[string]interface{})
	require.True(t, ok, "unlock result should be a map")
	
	// Verify unlock result contains expected fields
	unlockedAddress, ok := unlockResult["address"].(string)
	require.True(t, ok, "unlocked address should be a string")
	require.Equal(t, address, unlockedAddress, "unlocked address should match imported address")
	
	publicKey, ok := unlockResult["public_key"].(string)
	require.True(t, ok, "public key should be a string")
	require.NotEmpty(t, publicKey, "public key should not be empty")
	
	chains, ok := unlockResult["chains"].(map[string]interface{})
	require.True(t, ok, "chains should be a map")
	require.True(t, chains["ethereum"].(bool), "ethereum should be supported")
	
	unlockedAt, ok := unlockResult["unlocked_at"].(float64)
	require.True(t, ok, "unlocked_at should be a number")
	require.Greater(t, unlockedAt, float64(0), "unlocked_at should be a positive timestamp")

	// Verify wallet is now unlocked
	statusResponse2, err := nativeMsg.RpcRequest(ctx, "wallet_status", map[string]interface{}{})
	require.NoError(t, err, "failed to get wallet status after unlock")
	require.NotContains(t, statusResponse2, "error", "wallet status should not return error")
	
	statusResult2, ok := statusResponse2["result"].(map[string]interface{})
	require.True(t, ok, "status result should be a map")
	require.True(t, statusResult2["has_wallet"].(bool), "should have wallet")
	require.True(t, statusResult2["is_unlocked"].(bool), "wallet should be unlocked")
	require.Equal(t, address, statusResult2["address"].(string), "status should show correct address")
}

func TestUnlockWalletHandler_WrongPassword(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	nativeMsg := testEnv.GetNativeMsg()
	require.NotNil(t, nativeMsg, "Native messaging manager should not be nil")

	// First, import a wallet
	importParams := map[string]interface{}{
		"mnemonic":        "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		"password":        "CorrectPassword123!",
		"chain":           "ethereum",
		"derivationPath":  "m/44'/60'/0'/0/0",
	}

	importResponse, err := nativeMsg.RpcRequest(ctx, "import_wallet", importParams)
	require.NoError(t, err, "failed to import wallet")
	require.NotContains(t, importResponse, "error", "import wallet should not return error")

	// Wait for wallet to be saved
	time.Sleep(100 * time.Millisecond)

	// Lock the wallet
	_, err = nativeMsg.RpcRequest(ctx, "lock_wallet", map[string]interface{}{})
	require.NoError(t, err, "failed to lock wallet")

	// Try to unlock with wrong password
	unlockParams := map[string]interface{}{
		"password": "WrongPassword123!",
	}

	unlockResponse, err := nativeMsg.RpcRequest(ctx, "unlock_wallet", unlockParams)
	require.NoError(t, err, "RPC call should not fail")
	require.NotNil(t, unlockResponse, "unlock wallet response should not be nil")
	
	// Should return an error
	require.Contains(t, unlockResponse, "error", "unlock wallet should return error for wrong password")
	errorInfo, ok := unlockResponse["error"].(map[string]interface{})
	require.True(t, ok, "error should be a map")
	
	errorCode, ok := errorInfo["code"].(float64)
	require.True(t, ok, "error code should be a number")
	require.Equal(t, float64(-32001), errorCode, "should return invalid mnemonic/password error code")
	
	errorMessage, ok := errorInfo["message"].(string)
	require.True(t, ok, "error message should be a string")
	require.Contains(t, errorMessage, "Failed to unlock wallet", "error message should indicate unlock failure")

	// Verify wallet remains locked
	statusResponse, err := nativeMsg.RpcRequest(ctx, "wallet_status", map[string]interface{}{})
	require.NoError(t, err, "failed to get wallet status")
	require.NotContains(t, statusResponse, "error", "wallet status should not return error")
	
	statusResult, ok := statusResponse["result"].(map[string]interface{})
	require.True(t, ok, "status result should be a map")
	require.True(t, statusResult["has_wallet"].(bool), "should have wallet")
	require.False(t, statusResult["is_unlocked"].(bool), "wallet should remain locked")
}

func TestUnlockWalletHandler_NoWallet(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	nativeMsg := testEnv.GetNativeMsg()
	require.NotNil(t, nativeMsg, "Native messaging manager should not be nil")

	// First check if a wallet already exists
	statusResponse, err := nativeMsg.RpcRequest(ctx, "wallet_status", map[string]interface{}{})
	require.NoError(t, err, "failed to get wallet status")
	statusResult, ok := statusResponse["result"].(map[string]interface{})
	require.True(t, ok, "status result should be a map")
	
	hasWallet := statusResult["has_wallet"].(bool)
	
	if hasWallet {
		t.Skip("Wallet already exists in test environment, skipping no-wallet test")
	}

	// Try to unlock when no wallet exists
	unlockParams := map[string]interface{}{
		"password": "SomePassword123!",
	}

	unlockResponse, err := nativeMsg.RpcRequest(ctx, "unlock_wallet", unlockParams)
	require.NoError(t, err, "RPC call should not fail")
	require.NotNil(t, unlockResponse, "unlock wallet response should not be nil")
	
	// Should return an error
	require.Contains(t, unlockResponse, "error", "unlock wallet should return error when no wallet exists")
	errorInfo, ok := unlockResponse["error"].(map[string]interface{})
	require.True(t, ok, "error should be a map")
	
	errorCode, ok := errorInfo["code"].(float64)
	require.True(t, ok, "error code should be a number")
	require.Equal(t, float64(-32004), errorCode, "should return wallet not found error code")
	
	errorMessage, ok := errorInfo["message"].(string)
	require.True(t, ok, "error message should be a string")
	require.Contains(t, errorMessage, "No wallet found", "error message should indicate no wallet found")
}

func TestUnlockWalletHandler_MissingPassword(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	nativeMsg := testEnv.GetNativeMsg()
	require.NotNil(t, nativeMsg, "Native messaging manager should not be nil")

	// Try to unlock without password parameter
	unlockParams := map[string]interface{}{}

	unlockResponse, err := nativeMsg.RpcRequest(ctx, "unlock_wallet", unlockParams)
	require.NoError(t, err, "RPC call should not fail")
	require.NotNil(t, unlockResponse, "unlock wallet response should not be nil")
	
	// Should return a parameter error
	require.Contains(t, unlockResponse, "error", "unlock wallet should return error for missing password")
	errorInfo, ok := unlockResponse["error"].(map[string]interface{})
	require.True(t, ok, "error should be a map")
	
	errorCode, ok := errorInfo["code"].(float64)
	require.True(t, ok, "error code should be a number")
	require.Equal(t, float64(-32602), errorCode, "should return invalid params error code")
	
	errorMessage, ok := errorInfo["message"].(string)
	require.True(t, ok, "error message should be a string")
	require.Contains(t, errorMessage, "Password is required", "error message should indicate password is required")
}

func TestLockWalletHandler(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	nativeMsg := testEnv.GetNativeMsg()
	require.NotNil(t, nativeMsg, "Native messaging manager should not be nil")

	// Import and unlock a wallet first
	importParams := map[string]interface{}{
		"mnemonic":        "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		"password":        "TestPassword123!",
		"chain":           "ethereum",
		"derivationPath":  "m/44'/60'/0'/0/0",
	}

	_, err = nativeMsg.RpcRequest(ctx, "import_wallet", importParams)
	require.NoError(t, err, "failed to import wallet")

	// Verify wallet is unlocked after import
	statusResponse, err := nativeMsg.RpcRequest(ctx, "wallet_status", map[string]interface{}{})
	require.NoError(t, err, "failed to get wallet status")
	statusResult, ok := statusResponse["result"].(map[string]interface{})
	require.True(t, ok, "status result should be a map")
	require.True(t, statusResult["is_unlocked"].(bool), "wallet should be unlocked after import")

	// Now lock the wallet
	lockResponse, err := nativeMsg.RpcRequest(ctx, "lock_wallet", map[string]interface{}{})
	require.NoError(t, err, "failed to lock wallet")
	require.NotNil(t, lockResponse, "lock wallet response should not be nil")
	require.NotContains(t, lockResponse, "error", "lock wallet should not return error")
	require.Contains(t, lockResponse, "result", "lock wallet should return result")

	lockResult, ok := lockResponse["result"].(map[string]interface{})
	require.True(t, ok, "lock result should be a map")
	require.True(t, lockResult["locked"].(bool), "should indicate wallet is locked")

	// Verify wallet is now locked
	statusResponse2, err := nativeMsg.RpcRequest(ctx, "wallet_status", map[string]interface{}{})
	require.NoError(t, err, "failed to get wallet status after lock")
	statusResult2, ok := statusResponse2["result"].(map[string]interface{})
	require.True(t, ok, "status result should be a map")
	require.True(t, statusResult2["has_wallet"].(bool), "should still have wallet")
	require.False(t, statusResult2["is_unlocked"].(bool), "wallet should be locked")
	require.NotContains(t, statusResult2, "address", "address should not be present when locked")
}

func TestWalletStatusHandler(t *testing.T) {
	ctx := context.Background()
	testEnv, err := env.NewMcpHostTestEnvironment(nil)
	require.NoError(t, err, "failed to create test environment")
	defer testEnv.Cleanup()

	require.NoError(t, testEnv.Setup(ctx), "failed to setup test environment")

	nativeMsg := testEnv.GetNativeMsg()
	require.NotNil(t, nativeMsg, "Native messaging manager should not be nil")

	// Test getting wallet status (don't assume initial state)
	statusResponse, err := nativeMsg.RpcRequest(ctx, "wallet_status", map[string]interface{}{})
	require.NoError(t, err, "failed to get wallet status")
	require.NotContains(t, statusResponse, "error", "wallet status should not return error")
	
	statusResult, ok := statusResponse["result"].(map[string]interface{})
	require.True(t, ok, "status result should be a map")
	
	initialHasWallet := statusResult["has_wallet"].(bool)
	initialIsUnlocked := statusResult["is_unlocked"].(bool)
	
	t.Logf("Initial wallet status: has_wallet=%v, is_unlocked=%v", initialHasWallet, initialIsUnlocked)

	// Import a wallet
	importParams := map[string]interface{}{
		"mnemonic":        "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		"password":        "TestPassword123!",
		"chain":           "ethereum",
		"derivationPath":  "m/44'/60'/0'/0/0",
	}

	importResponse, err := nativeMsg.RpcRequest(ctx, "import_wallet", importParams)
	require.NoError(t, err, "failed to import wallet")
	importResult, ok := importResponse["result"].(map[string]interface{})
	require.True(t, ok, "import result should be a map")
	address := importResult["address"].(string)

	// Test status after wallet import (should be unlocked)
	statusResponse2, err := nativeMsg.RpcRequest(ctx, "wallet_status", map[string]interface{}{})
	require.NoError(t, err, "failed to get wallet status after import")
	statusResult2, ok := statusResponse2["result"].(map[string]interface{})
	require.True(t, ok, "status result should be a map")
	require.True(t, statusResult2["has_wallet"].(bool), "should have wallet after import")
	require.True(t, statusResult2["is_unlocked"].(bool), "should be unlocked after import")
	require.Equal(t, address, statusResult2["address"].(string), "should show correct address when unlocked")

	// Lock the wallet
	_, err = nativeMsg.RpcRequest(ctx, "lock_wallet", map[string]interface{}{})
	require.NoError(t, err, "failed to lock wallet")

	// Test status after locking
	statusResponse3, err := nativeMsg.RpcRequest(ctx, "wallet_status", map[string]interface{}{})
	require.NoError(t, err, "failed to get wallet status after lock")
	statusResult3, ok := statusResponse3["result"].(map[string]interface{})
	require.True(t, ok, "status result should be a map")
	require.True(t, statusResult3["has_wallet"].(bool), "should still have wallet after lock")
	require.False(t, statusResult3["is_unlocked"].(bool), "should not be unlocked after lock")
	require.NotContains(t, statusResult3, "address", "address should not be present when locked")
}