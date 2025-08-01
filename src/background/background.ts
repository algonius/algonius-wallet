// Algonius Wallet - Background Service Worker (TypeScript)
// MCP-controlled multi-chain trading wallet with Native Host integration

import { McpHostManager } from '../mcp/McpHostManager';

// Initialize MCP Host Manager
const mcpHostManager = new McpHostManager();

// Extension lifecycle events
chrome.runtime.onInstalled.addListener(async () => {
  console.log('Algonius Wallet installed');
  // Try to start MCP Host with default configuration
  try {
    await mcpHostManager.startMcpHost({
      runMode: 'development',
      logLevel: 'info'
    });
  } catch (error) {
    console.log('Failed to start MCP Host on install:', error);
  }
});

chrome.runtime.onStartup.addListener(async () => {
  console.log('Algonius Wallet startup');
  // Try to start MCP Host with default configuration
  try {
    await mcpHostManager.startMcpHost({
      runMode: 'development',
      logLevel: 'info'
    });
  } catch (error) {
    console.log('Failed to start MCP Host on startup:', error);
  }
});

/**
 * Message handling for popup and other extension components
 */
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (typeof request === "object" && request !== null && "action" in request) {
    switch (request.action) {
      case "get_mcp_status":
        sendResponse({ status: mcpHostManager.getStatus() });
        return true;
        
      case "connect_mcp":
        mcpHostManager.connect()
          .then(connected => {
            sendResponse({ success: connected });
          })
          .catch(error => {
            sendResponse({ success: false, error: error.message });
          });
        return true;
        
      case "start_mcp":
        mcpHostManager.startMcpHost({
          runMode: 'development',
          logLevel: 'info'
        })
          .then(started => {
            sendResponse({ success: started });
          })
          .catch(error => {
            sendResponse({ success: false, error: error.message });
          });
        return true;
        
      case "stop_mcp":
        mcpHostManager.stopMcpHost()
          .then(stopped => {
            sendResponse({ success: stopped });
          })
          .catch(error => {
            sendResponse({ success: false, error: error.message });
          });
        return true;
        
      case "native_rpc":
        // Handle native messaging RPC requests
        handleNativeRpc(request, sender, sendResponse);
        return true;
    }
  }
  
  // Handle content script messages for Web3 provider functionality  
  if (typeof request === "object" && request !== null && "type" in request && request.type === "ALGONIUS_WALLET_FORWARD") {
    handleWeb3Request(request.data, sender, sendResponse);
    return true;
  }
  
  // Handle direct Web3 requests (legacy support)
  if (typeof request === "object" && request !== null && "method" in request) {
    handleWeb3Request(request, sender, sendResponse);
    return true;
  }
  
  // Keep service worker active
  return true;
});

/**
 * Handle native messaging RPC requests from popup
 */
async function handleNativeRpc(
  request: unknown,
  sender: chrome.runtime.MessageSender,
  sendResponse: (response?: unknown) => void
) {
  try {
    console.log('Native RPC:', (request as { method?: string })?.method);
    
    if (!mcpHostManager.getStatus().isConnected) {
      console.log('❌ MCP Host not connected');
      sendResponse({ error: { message: 'MCP Host not connected' } });
      return;
    }

    if (
      typeof request !== "object" ||
      request === null ||
      !("method" in request) ||
      !("params" in request)
    ) {
      console.log('❌ Invalid RPC request format');
      sendResponse({ error: { message: "Invalid native RPC request" } });
      return;
    }

    const { method, params } = request as { method: string; params: unknown };
    
    // Forward the request to MCP Host via RPC
    const response = await mcpHostManager.rpcRequest({
      method,
      params
    });
    
    if (response.error) {
      console.log('❌ RPC Error:', response.error);
      sendResponse({ error: response.error });
    } else {
      console.log('✅ RPC Success:', method, 'Response:', response.result);
      sendResponse({ result: response.result });
    }
  } catch (error) {
    console.error('Native RPC request failed:', error);
    sendResponse({ 
      error: { 
        code: -32603,
        message: error instanceof Error ? error.message : String(error) 
      } 
    });
  }
}

/** 
 * Handle Web3 provider requests from content scripts
 */
async function handleWeb3Request(
  request: unknown,
  sender: chrome.runtime.MessageSender,
  sendResponse: (response?: unknown) => void
) {
  try {
    console.log('Handling Web3 request:', request);
    
    const status = mcpHostManager.getStatus();
    console.log('MCP Host status:', status);
    
    if (!status.isConnected) {
      console.log('MCP Host not connected');
      sendResponse({ error: 'MCP Host not connected' });
      return;
    }

    if (
      typeof request !== "object" ||
      request === null ||
      !("method" in request)
    ) {
      console.log('Invalid Web3 request');
      sendResponse({ error: "Invalid Web3 request" });
      return;
    }

    // Forward the request to MCP Host via RPC
    console.log('Forwarding request to MCP Host');
    const response = await mcpHostManager.rpcRequest({
      method: 'web3_request',
      params: {
        method: (request as { method: string }).method,
        params: (request as { params?: unknown }).params,
        origin: sender.tab?.url
      }
    });
    
    console.log('Web3 request response from MCP Host:', response);
    // 确保响应格式正确
    sendResponse({
      result: response.result,
      error: response.error
    });
  } catch (error) {
    console.error('Web3 request failed:', error);
    sendResponse({ 
      error: error instanceof Error ? error.message : String(error) 
    });
  }
}

// Periodic wake-up
setInterval(() => {
  console.log('Algonius Wallet heartbeat');
}, 25000);
