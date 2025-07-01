import React, { useEffect, useState, useCallback } from "react";

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

  // 获取MCP Host状态
  const fetchStatus = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await getMcpStatus();
      setStatus(res.status);
    } catch (err) {
      setError('Failed to get MCP Host status');
      console.error('Failed to get MCP status:', err);
    }
    setLoading(false);
  }, []);

  // 首次加载和定时刷新
  useEffect(() => {
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

  return (
    <div className="w-80 max-w-xs mx-auto bg-white rounded shadow p-4 min-h-screen font-sans">
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

      <main>
        <section className="mb-4">
          <h2 className="text-lg font-semibold mb-2">Portfolio</h2>
          <div id="balance-display" className="bg-gray-100 rounded p-2 text-center">
            <div className="text-gray-400">
              {status.isConnected ? "Loading balances..." : "MCP Host not connected"}
            </div>
          </div>
        </section>

        <section className="mb-4">
          <h2 className="text-lg font-semibold mb-2">Active Signals</h2>
          <div id="signals-list" className="bg-gray-100 rounded p-2 text-center">
            <div className="text-gray-400">
              {status.isConnected ? "Loading signals..." : "MCP Host not connected"}
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
};

export default App;
