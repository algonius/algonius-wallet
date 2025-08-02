import React, { useCallback, useEffect, useState } from "react";
import { CreateWallet } from "./components/WalletSetup/CreateWallet";
import { ImportWallet } from "./components/WalletSetup/ImportWallet";
import { UnlockWallet } from "./components/WalletSetup/UnlockWallet";
import { Button } from "./components/common/Button";
import { useNativeMessaging } from "./hooks/useNativeMessaging";

// MCP Host Status interface
interface McpHostStatus {
  isConnected: boolean;
  startTime: number | null;
  lastHeartbeat: number | null;
  version: string | null;
  runMode: string | null;
  uptime?: string;
  ssePort?: string;
  sseBaseURL?: string;
}

// App view states
type AppView = 'main' | 'setup' | 'create' | 'import' | 'unlock' | 'wallet-ready';

const getMcpStatus = (): Promise<{ status: McpHostStatus }> => {
  return new Promise((resolve) => {
    chrome.runtime.sendMessage({ action: "get_mcp_status" }, (response) => {
      resolve(response || { status: { isConnected: false, startTime: null, lastHeartbeat: null, version: null, runMode: null } });
    });
  });
};

const connectMcp = (): Promise<{ success: boolean; error?: string }> => {
  return new Promise((resolve) => {
    chrome.runtime.sendMessage({ action: "connect_mcp" }, (response) => {
      resolve(response || { success: false });
    });
  });
};

const startMcp = (): Promise<{ success: boolean; error?: string }> => {
  return new Promise((resolve) => {
    chrome.runtime.sendMessage({ action: "start_mcp" }, (response) => {
      resolve(response || { success: false });
    });
  });
};

const stopMcp = (): Promise<{ success: boolean; error?: string }> => {
  return new Promise((resolve) => {
    chrome.runtime.sendMessage({ action: "stop_mcp" }, (response) => {
      resolve(response || { success: false });
    });
  });
};

const App: React.FC = () => {
  const [status, setStatus] = useState<McpHostStatus>({
    isConnected: false,
    startTime: null,
    lastHeartbeat: null,
    version: null,
    runMode: null,
  });
  const [loading, setLoading] = useState<boolean>(true);
  const [actionLoading, setActionLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [currentView, setCurrentView] = useState<AppView>('main');
  const [walletStatus, setWalletStatus] = useState<{
    hasWallet: boolean;
    isUnlocked: boolean;
    address?: string;
  }>({ hasWallet: false, isUnlocked: false });

  const { getWalletStatus } = useNativeMessaging();

  // 获取MCP Host状态
  const fetchStatus = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await getMcpStatus();
      setStatus(res.status);
      
      // 如果MCP Host已连接，检查钱包状态
      if (res.status.isConnected) {
        try {
          const walletRes = await getWalletStatus();
          setWalletStatus(walletRes);
          
          // 如果钱包已存在且当前视图是status，并且用户还没有手动切换到wallet视图，切换到wallet视图
          // 但只在用户第一次打开popup时自动切换，避免在用户手动导航时干扰
          // 同时确保钱包确实是解锁状态
          if (walletRes.hasWallet && walletRes.isUnlocked && 
              !sessionStorage.getItem('autoWalletRedirectDone')) {
            setCurrentView('main');
            // 标记已执行过自动重定向
            sessionStorage.setItem('autoWalletRedirectDone', 'true');
          }
          // 如果钱包已解锁且当前在wallet视图，确保界面正确显示
          else if (walletRes.hasWallet && walletRes.isUnlocked && currentView === 'wallet-ready') {
            // 强制更新状态以确保界面正确显示
            setWalletStatus(prev => ({...prev, ...walletRes}));
          }
        } catch (walletErr) {
          console.log('Wallet status check failed:', walletErr);
          // 钱包状态检查失败不应该影响MCP状态显示
        }
      }
    } catch (err) {
      setError('Failed to get MCP Host status');
      console.error('Failed to get MCP status:', err);
    }
    setLoading(false);
  }, [getWalletStatus, currentView]);

  // 首次加载和定时刷新
  useEffect(() => {
    // 清除自动重定向标记，确保每次打开popup都能正确显示
    sessionStorage.removeItem('autoWalletRedirectDone');
    fetchStatus();
    const interval = setInterval(fetchStatus, 3000);
    return () => clearInterval(interval);
  }, [fetchStatus]);

  // 监听状态更新消息
  useEffect(() => {
    const handleMessage = (message: unknown) => {
      if (
        typeof message === "object" &&
        message !== null &&
        "type" in message &&
        (message as { type?: unknown }).type === "mcpHostStatusUpdate" &&
        "status" in message
      ) {
        setStatus((message as { status: McpHostStatus }).status);
      }
    };

    chrome.runtime.onMessage.addListener(handleMessage);
    return () => chrome.runtime.onMessage.removeListener(handleMessage);
  }, []);

  // 连接MCP Host
  const handleConnect = async () => {
    setActionLoading(true);
    setError(null);
    try {
      const result = await connectMcp();
      if (!result.success) {
        setError(result.error || 'Failed to connect to MCP Host');
      }
    } catch (err) {
      setError('Failed to connect to MCP Host');
      console.error('Connect failed:', err);
    }
    setActionLoading(false);
    // Refresh status after action
    setTimeout(fetchStatus, 500);
  };

  // 启动MCP Host
  const handleStart = async () => {
    setActionLoading(true);
    setError(null);
    try {
      const result = await startMcp();
      if (!result.success) {
        setError(result.error || 'Failed to start MCP Host');
      }
    } catch (err) {
      setError('Failed to start MCP Host');
      console.error('Start failed:', err);
    }
    setActionLoading(false);
    // Refresh status after action
    setTimeout(fetchStatus, 500);
  };

  // 停止MCP Host
  const handleStop = async () => {
    setActionLoading(true);
    setError(null);
    try {
      const result = await stopMcp();
      if (!result.success) {
        setError(result.error || 'Failed to stop MCP Host');
      }
    } catch (err) {
      setError('Failed to stop MCP Host');
      console.error('Stop failed:', err);
    }
    setActionLoading(false);
    // Refresh status after action
    setTimeout(fetchStatus, 500);
  };

  // 格式化时间显示
  const formatTime = (timestamp: number | null): string => {
    if (!timestamp) return '--';
    return new Date(timestamp).toLocaleTimeString();
  };

  // 计算运行时间
  const getUptime = (): string => {
    if (!status.startTime) return '--';
    const now = Date.now();
    const uptime = now - status.startTime;
    const minutes = Math.floor(uptime / 60000);
    const seconds = Math.floor((uptime % 60000) / 1000);
    return `${minutes}m ${seconds}s`;
  };

  // View management functions
  const handleSetupWallet = useCallback(() => {
    // If wallet exists but is locked, go to unlock view
    // Otherwise go to setup view
    if (walletStatus.hasWallet && !walletStatus.isUnlocked) {
      setCurrentView('unlock');
    } else {
      setCurrentView('setup');
    }
  }, [walletStatus]);

  const handleCreateWallet = useCallback(() => {
    setCurrentView('create');
  }, []);

  const handleImportWallet = useCallback(() => {
    setCurrentView('import');
  }, []);

  const handleWalletComplete = useCallback(async () => {
    // 钱包设置完成后，刷新钱包状态
    try {
      const walletRes = await getWalletStatus();
      setWalletStatus(walletRes);
      setCurrentView('wallet-ready');
    } catch (err) {
      console.error('Failed to refresh wallet status:', err);
      setCurrentView('wallet-ready'); // 仍然切换到钱包视图
    }
  }, [getWalletStatus]);

  const handleBackToStatus = useCallback(() => {
    setCurrentView('main');
  }, []);

  const handleUnlockComplete = useCallback(async () => {
    // 钱包解锁完成后，刷新钱包状态并进入wallet视图
    try {
      const walletRes = await getWalletStatus();
      setWalletStatus(walletRes);
      // 标记已执行过自动重定向，避免定时刷新时再次自动切换
      sessionStorage.setItem('autoWalletRedirectDone', 'true');
      // 直接进入wallet视图而不是status视图
      setCurrentView('wallet-ready');
      // 同时更新MCP状态以确保界面同步
      fetchStatus();
    } catch (err) {
      console.error('Failed to refresh wallet status:', err);
      // 出错时也进入wallet视图
      setCurrentView('wallet-ready');
    }
  }, [getWalletStatus, fetchStatus]);

  // Render different views based on current view state
  const renderView = () => {
    switch (currentView) {
      case 'setup':
        return (
          <div className="space-y-6">
            <div className="text-center space-y-4">
              <div className="mx-auto w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center">
                <svg className="w-8 h-8 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
                </svg>
              </div>
              <div>
                <h2 className="text-xl font-semibold text-gray-900">Setup Your Wallet</h2>
                <p className="text-sm text-gray-600 mt-2">
                  Create a new wallet or import an existing one
                </p>
              </div>
            </div>

            <div className="space-y-3">
              <Button
                variant="primary"
                fullWidth
                onClick={handleCreateWallet}
              >
                Create New Wallet
              </Button>
              
              <Button
                variant="secondary"
                fullWidth
                onClick={handleImportWallet}
              >
                Import Existing Wallet
              </Button>
            </div>

            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <h3 className="text-sm font-medium text-blue-800 mb-2">Getting Started</h3>
              <ul className="text-sm text-blue-700 space-y-1">
                <li>• New users should create a new wallet</li>
                <li>• Existing users can import with their recovery phrase</li>
                <li>• Your wallet will be secured with a password</li>
              </ul>
            </div>

            <div className="flex justify-center">
              <button
                onClick={handleBackToStatus}
                className="text-sm text-gray-500 hover:text-gray-700"
              >
                Back to Status
              </button>
            </div>
          </div>
        );

      case 'create':
        return (
          <CreateWallet
            onComplete={handleWalletComplete}
            onCancel={handleBackToStatus}
          />
        );

      case 'import':
        return (
          <ImportWallet
            onComplete={handleWalletComplete}
            onCancel={handleBackToStatus}
          />
        );

      case 'unlock':
        return (
          <UnlockWallet
            onComplete={handleUnlockComplete}
            onCancel={handleBackToStatus}
          />
        );

      case 'wallet-ready':
        return (
          <div className="space-y-6">
            <div className="text-center space-y-4">
              <div className="mx-auto w-16 h-16 bg-green-100 rounded-full flex items-center justify-center">
                <svg className="w-8 h-8 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>
              <div>
                <h2 className="text-xl font-semibold text-gray-900">Wallet Ready</h2>
                <p className="text-sm text-gray-600 mt-2">
                  Your wallet is set up and ready to use
                </p>
              </div>
            </div>

            <div className="bg-green-50 border border-green-200 rounded-lg p-4">
              <h3 className="text-sm font-medium text-green-800 mb-2">Next Steps</h3>
              <ul className="text-sm text-green-700 space-y-1">
                <li>• Your wallet is now connected to the MCP Host</li>
                <li>• You can now interact with dApps</li>
                <li>• AI agents can manage your transactions</li>
              </ul>
            </div>

            <Button
              variant="primary"
              fullWidth
              onClick={handleBackToStatus}
            >
              Go to Wallet Dashboard
            </Button>
          </div>
        );

      default: // 'status'
        return (
          <div>
            {/* Error Display */}
            {error && (
              <div className="mb-4 bg-red-50 border border-red-200 rounded p-2">
                <span className="text-xs text-red-700">{error}</span>
                <button
                  className="ml-2 text-red-500 hover:text-red-700"
                  onClick={() => setError(null)}
                >
                  ×
                </button>
              </div>
            )}

            {/* MCP Host Status Section */}
            <section className="mb-4">
              <h2 className="text-lg font-semibold mb-2">MCP Host Status</h2>
              <div className="bg-gray-100 rounded p-3 space-y-2">
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Connection:</span>
                  <span className={`text-sm ${status.isConnected ? 'text-green-600' : 'text-red-600'}`}>
                    {status.isConnected ? 'Connected' : 'Disconnected'}
                  </span>
                </div>
                
                {status.isConnected && (
                  <>
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium">Version:</span>
                      <span className="text-sm text-gray-600">{status.version || '--'}</span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium">Uptime:</span>
                      <span className="text-sm text-gray-600">{getUptime()}</span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium">Last Heartbeat:</span>
                      <span className="text-sm text-gray-600">{formatTime(status.lastHeartbeat)}</span>
                    </div>
                    {status.ssePort && (
                      <div className="flex justify-between items-center">
                        <span className="text-sm font-medium">SSE Port:</span>
                        <span className="text-sm text-gray-600">{status.ssePort}</span>
                      </div>
                    )}
                  </>
                )}
                
                <div className="flex space-x-2 mt-3">
                  {!status.isConnected ? (
                    <button
                      className="flex-1 bg-blue-500 text-white px-3 py-1 rounded text-sm hover:bg-blue-600 disabled:opacity-50"
                      onClick={handleStart}
                      disabled={actionLoading || loading}
                    >
                      {actionLoading ? "Starting..." : "Start MCP Host"}
                    </button>
                  ) : (
                    <button
                      className="flex-1 bg-red-500 text-white px-3 py-1 rounded text-sm hover:bg-red-600 disabled:opacity-50"
                      onClick={handleStop}
                      disabled={actionLoading || loading}
                    >
                      {actionLoading ? "Stopping..." : "Stop MCP Host"}
                    </button>
                  )}
                  <button
                    className="bg-gray-500 text-white px-3 py-1 rounded text-sm hover:bg-gray-600 disabled:opacity-50"
                    onClick={handleConnect}
                    disabled={actionLoading || loading || status.isConnected}
                  >
                    {actionLoading ? "Connecting..." : "Reconnect"}
                  </button>
                </div>
              </div>
            </section>

            {/* Wallet Status Section */}
            <section className="mb-4">
              <h2 className="text-lg font-semibold mb-2">Wallet</h2>
              <div className="bg-gray-100 rounded p-3 space-y-2">
                {status.isConnected ? (
                  walletStatus.hasWallet ? (
                    <div className="text-center">
                      <div className="text-green-600 mb-2">
                        ✓ Wallet configured
                      </div>
                      {walletStatus.address && (
                        <div className="text-xs text-gray-600 mb-3 font-mono">
                          {walletStatus.address.slice(0, 6)}...{walletStatus.address.slice(-4)}
                        </div>
                      )}
                      <div className="flex justify-center space-x-2">
                        {walletStatus.isUnlocked ? (
                          <Button
                            variant="primary"
                            size="small"
                            onClick={() => {
                              // 标记已执行过自动重定向，避免定时刷新时再次自动切换
                              sessionStorage.setItem('autoWalletRedirectDone', 'true');
                              setCurrentView('wallet-ready');
                            }}
                          >
                            Go to Wallet
                          </Button>
                        ) : (
                          <Button
                            variant="secondary"
                            size="small"
                            onClick={handleSetupWallet}
                          >
                            Unlock Wallet
                          </Button>
                        )}
                      </div>
                    </div>
                  ) : (
                    <div className="text-center">
                      <div className="text-gray-400 mb-3">
                        No wallet configured
                      </div>
                      <Button
                        variant="primary"
                        size="small"
                        onClick={handleSetupWallet}
                      >
                        Setup Wallet
                      </Button>
                    </div>
                  )
                ) : (
                  <div className="text-center">
                    <div className="text-gray-400 mb-3">
                      MCP Host not connected
                    </div>
                  </div>
                )}
              </div>
            </section>

            <main>
              <section className="mb-4">
                <h2 className="text-lg font-semibold mb-2">Portfolio</h2>
                <div id="balance-display" className="bg-gray-100 rounded p-2 text-center">
                  <div className="text-gray-400">
                    {status.isConnected ? "Setup wallet to view balances" : "MCP Host not connected"}
                  </div>
                </div>
              </section>

              <section className="mb-4">
                <h2 className="text-lg font-semibold mb-2">Active Signals</h2>
                <div id="signals-list" className="bg-gray-100 rounded p-2 text-center">
                  <div className="text-gray-400">
                    {status.isConnected ? "Setup wallet to view signals" : "MCP Host not connected"}
                  </div>
                </div>
              </section>

              <section className="flex justify-between mb-4">
                <button 
                  id="refresh-btn" 
                  className="bg-blue-500 text-white px-3 py-1 rounded hover:bg-blue-600 disabled:opacity-50"
                  onClick={fetchStatus}
                  disabled={loading}
                >
                  {loading ? "Refreshing..." : "Refresh"}
                </button>
                <button id="settings-btn" className="bg-gray-200 text-gray-700 px-3 py-1 rounded hover:bg-gray-300">Settings</button>
              </section>
            </main>

            <footer className="flex justify-between items-center text-xs text-gray-400 border-t pt-2">
              <div className="version">v1.0.0</div>
              <div className="connection-info">
                <span id="last-updated">
                  Last updated: {formatTime(status.lastHeartbeat)}
                </span>
              </div>
            </footer>
          </div>
        );
    }
  };

  return (
    <div className="w-96 max-w-md mx-auto bg-white rounded shadow p-4 min-h-screen font-sans">
      <header className="flex items-center justify-between mb-4">
        <h1 className="text-xl font-bold">Algonius Wallet</h1>
        <div className="flex items-center space-x-2">
          <span
            id="connection-status"
            className={`w-2 h-2 rounded-full ${
              status.isConnected ? "bg-green-500" : "bg-gray-400"
            } transition-colors`}
            title={status.isConnected ? "MCP Host Connected" : "MCP Host Disconnected"}
          ></span>
          <span id="network-name" className="text-sm text-gray-600">
            {status.runMode || 'MCP Host'}
          </span>
        </div>
      </header>

      {renderView()}
    </div>
  );
};

export default App;
