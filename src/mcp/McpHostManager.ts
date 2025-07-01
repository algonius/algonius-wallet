/**
 * MCP Host Manager
 *
 * This class manages the connection to the MCP Host process, monitors its status,
 * and provides methods to control it (start, stop, etc.).
 * It also supports bidirectional RPC communication with the MCP Host.
 */
import { McpErrorCode, createMcpError } from '../shared';

// Define RPC types
export interface RpcRequest {
  /**
   * Unique identifier for the request
   */
  id?: string;

  /**
   * Method to be invoked
   */
  method: string;

  /**
   * Parameters for the method
   */
  params?: unknown;
}

/**
 * JSON-RPC like response structure
 */
export interface RpcResponse {
  /**
   * Identifier matching the request
   */
  id?: string;

  /**
   * Result of the method call (if successful)
   */
  result?: unknown;

  /**
   * Error information (if the call failed)
   */
  error?: {
    code: number;
    message: string;
    data?: unknown;
  };
}

/**
 * Options for RPC requests
 */
export interface RpcRequestOptions {
  /**
   * Timeout in milliseconds
   */
  timeout?: number;
}

/**
 * A function that handles an RPC request and returns a promise of RpcResponse
 */
export type RpcHandler = (request: RpcRequest) => Promise<RpcResponse>;

// Define the MCP Host status interface
export interface McpHostStatus {
  isConnected: boolean;
  startTime: number | null;
  lastHeartbeat: number | null;
  version: string | null;
  runMode: string | null;
  uptime?: string;
  ssePort?: string;
  sseBaseURL?: string;
}

// Define the MCP Host configuration options
export interface McpHostOptions {
  runMode: string;
  port?: number;
  logLevel?: string;
}

// Type for status change event listeners
export type StatusListener = (status: McpHostStatus) => void;

export class McpHostManager {
  private port: chrome.runtime.Port | null = null;
  private status: McpHostStatus = {
    isConnected: false,
    startTime: null,
    lastHeartbeat: null,
    version: null,
    runMode: null,
  };
  private listeners: StatusListener[] = [];
  private heartbeatInterval: NodeJS.Timeout | null = null;
  private pingTimeout: NodeJS.Timeout | null = null;
  private readonly HEARTBEAT_INTERVAL_MS = 10000; // 10 seconds
  private readonly PING_TIMEOUT_MS = 20000; // 20 seconds
  private readonly GRACEFUL_SHUTDOWN_TIMEOUT_MS = 1000; // 1 second
  // RPC-related properties
  private rpcMethodHandlers: Map<string, RpcHandler> = new Map();
  private pendingRequests = new Map<
    string,
    {
      resolve: (value: unknown) => void;
      reject: (reason?: unknown) => void;
      timeoutId: ReturnType<typeof setTimeout>;
    }
  >();
  private readonly RPC_TIMEOUT_MS = 5000; // 5 seconds default timeout for RPC requests

  /**
   * Establishes a connection to the MCP Host Native Messaging host.
   * @returns {Promise<boolean>} Promise that resolves to true if connection was established successfully
   * @throws Will reject with an error if connection fails (e.g., host not installed)
   */
  public connect(): Promise<boolean> {
    // Don't reconnect if already connected
    if (this.port) {
      return Promise.resolve(false);
    }

    return new Promise((resolve, reject) => {
      // Connect to the native messaging host
      // Note: connectNative always returns a Port object even if the host doesn't exist
      this.port = chrome.runtime.connectNative('ai.algonius.wallet');

      // Setup message handler
      this.port.onMessage.addListener(this.handleMessage.bind(this));

      // Critical: The onDisconnect event is where we need to check for connection errors
      // We'll create a local handler first that includes the Promise resolution
      const disconnectHandler = () => {
        // Check for runtime error which indicates connection failure
        const lastError = chrome.runtime.lastError;
        if (lastError) {
          const errorMessage = lastError.message;
          console.error(`Native messaging connection failed: ${errorMessage}`);

          // Clean up port reference
          this.port = null;

          // Create a structured MCP error with appropriate error code
          const mcpError = createMcpError(
            McpErrorCode.HOST_NOT_FOUND,
            `Native messaging connection failed: ${errorMessage}`,
            { originalError: lastError.message },
          );

          // Reject the promise with the structured error
          reject(mcpError);
        } else {
          // This was a normal disconnect - should be handled by the main disconnect handler
          this.handleDisconnect();
        }
      };

      // Add our disconnect handler
      this.port.onDisconnect.addListener(disconnectHandler);

      // Set a short timeout to verify successful connection
      // Most connection errors trigger onDisconnect almost immediately
      setTimeout(() => {
        if (this.port) {
          // Update and broadcast status
          this.updateStatus({ isConnected: true });

          // Start heartbeat
          this.startHeartbeat();

          // Resolve the promise
          resolve(true);

          this.port.onDisconnect.removeListener(disconnectHandler);
        }
      }, 100);
    });
  }

  /**
   * Disconnects from the MCP Host.
   */
  public disconnect(): void {
    if (!this.port) {
      return;
    }

    // Clear heartbeat timer
    this.stopHeartbeat();

    // Disconnect
    this.port.disconnect();
    this.port = null;

    // Update status
    this.updateStatus({
      isConnected: false,
      lastHeartbeat: null,
    });
  }

  /**
   * Gets the current MCP Host status.
   * @returns {McpHostStatus} The current status.
   */
  public getStatus(): McpHostStatus {
    return { ...this.status };
  }

  /**
   * Registers a listener for status changes.
   * @param {StatusListener} listener The listener to add.
   */
  public addStatusListener(listener: StatusListener): void {
    this.listeners.push(listener);
  }

  /**
   * Removes a status change listener.
   * @param {StatusListener} listener The listener to remove.
   */
  public removeStatusListener(listener: StatusListener): void {
    const index = this.listeners.indexOf(listener);
    if (index !== -1) {
      this.listeners.splice(index, 1);
    }
  }

  /**
   * Starts the MCP Host process using RPC.
   * @param {McpHostOptions} options Configuration options.
   * @returns {Promise<boolean>} True if the Host was started successfully.
   * @throws Will reject with an error if connection fails (e.g., host not installed)
   */
  public async startMcpHost(options: McpHostOptions): Promise<boolean> {
    // Don't start if already connected
    if (this.status.isConnected) {
      return false;
    }

    try {
      // First establish connection to MCP Host
      const connected = await this.connect();
      if (!connected) {
        return false;
      }

      // Send init RPC request
      const response = await this.rpcRequest({
        method: 'init',
        params: options,
      });

      // Check if init was successful
      if (
        response &&
        typeof response.result === "object" &&
        response.result !== null &&
        "status" in response.result &&
        (response.result as { status?: unknown }).status === "initialized"
      ) {
        console.log('MCP Host initialized successfully');
        // Set the start time when the host is successfully initialized
        this.updateStatus({ startTime: Date.now() });
        return true;
      } else {
        console.error('MCP Host init failed:', response.error || 'Unknown error');
        return false;
      }
    } catch (error) {
      console.error('Failed to start MCP Host:', error);
      throw error;
    }
  }

  /**
   * Stops the MCP Host process using RPC.
   * @returns {Promise<boolean>} True if the Host was stopped successfully.
   */
  public async stopMcpHost(): Promise<boolean> {
    if (!this.port || !this.status.isConnected) {
      return false;
    }

    try {
      // Send shutdown RPC request
      const response = await this.rpcRequest(
        {
          method: 'shutdown',
          params: {},
        },
        { timeout: this.GRACEFUL_SHUTDOWN_TIMEOUT_MS },
      );

      // Check if shutdown was successful
      if (
        response &&
        typeof response.result === "object" &&
        response.result !== null &&
        "status" in response.result &&
        (response.result as { status?: unknown }).status === "shutting_down"
      ) {
        console.log('MCP Host shutdown initiated');

        // Wait for disconnect event or timeout
        return new Promise(resolve => {
          const timeout = setTimeout(() => {
            // Force disconnect if graceful shutdown times out
            this.disconnect();
            resolve(true);
          }, this.GRACEFUL_SHUTDOWN_TIMEOUT_MS);

          // Add a one-time disconnect listener
          const disconnectHandler = () => {
            const lastError = chrome.runtime.lastError;
            if (lastError) {
              const errorMessage = lastError.message;
              console.info(`Native messaging disconnect ok, message: ${errorMessage}`);
              this.port = null;
            }

            clearTimeout(timeout);

            // Update status
            this.updateStatus({
              isConnected: false,
              lastHeartbeat: null,
              startTime: null,
            });

            resolve(true);
          };

          // Wait for disconnect event
          this.port?.onDisconnect.addListener(disconnectHandler);
        });
      } else {
        console.error('MCP Host shutdown failed:', response.error || 'Unknown error');
        return false;
      }
    } catch (error) {
      console.error('Failed to stop MCP Host:', error);
      // Force disconnect on error
      this.disconnect();
      return true;
    }
  }

  /**
   * Registers an RPC method handler that can be called by the MCP Host.
   * @param method The RPC method name to register
   * @param handler The function to handle requests for this method
   */
  public registerRpcMethod(method: string, handler: RpcHandler): void {
    console.debug(`[McpHostManager] Registering RPC handler for method: ${method}`);
    this.rpcMethodHandlers.set(method, handler);
  }

  /**
   * Sends an RPC request to the MCP Host and returns a promise for the response.
   * @param rpc The RPC request to send
   * @param options Optional configuration for the request
   * @returns A Promise that resolves with the response or rejects on error/timeout
   */
  public rpcRequest(rpc: RpcRequest, options: RpcRequestOptions = {}): Promise<RpcResponse> {
    if (!this.port || !this.status.isConnected) {
      return Promise.reject(new Error('Cannot send RPC request: host not connected'));
    }

    // Generate a unique ID if not provided
    const id = rpc.id ?? this.generateRequestId();
    const method = rpc.method;
    const params = rpc.params;

    // Get timeout from options or use default
    const { timeout = this.RPC_TIMEOUT_MS } = options;

    console.debug(`[McpHostManager] Sending RPC request: ${method} (id: ${id})`);

    return new Promise<unknown>((resolve, reject) => {
      // Set up timeout to reject the promise if no response is received
      const timeoutId = setTimeout(() => {
        this.pendingRequests.delete(id);
        reject(new Error(`RPC request timeout: ${method} (id: ${id})`));
      }, timeout);

      // Store the promise resolvers with the request ID
      this.pendingRequests.set(id, {
        resolve: (value: unknown) => resolve(value),
        reject,
        timeoutId
      });

      // Send the RPC request message
      this.port?.postMessage({
        type: 'rpc_request',
        id,
        method,
        params,
      });
    }) as Promise<RpcResponse>;
  }

  /**
   * Generates a unique ID for RPC requests
   * @returns A unique string ID
   */
  private generateRequestId(): string {
    return 'req_' + Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
  }

  /**
   * Processes an incoming RPC request from the MCP Host.
   * @param data The RPC request data
   */
  private async handleRpcRequest(data: unknown): Promise<void> {
    if (
      typeof data !== "object" ||
      data === null ||
      !("method" in data) ||
      !("id" in data)
    ) {
      console.warn("[McpHostManager] Invalid RPC request data", data);
      return;
    }
    const { method, id, params } = data as { method: string; id?: string; params?: unknown };
    console.log(`[McpHostManager] Received RPC request: ${method} (id: ${id})`);

    // Find the registered handler for this method
    const handler = this.rpcMethodHandlers.get(method);

    if (!handler) {
      console.warn(`[McpHostManager] No handler registered for RPC method: ${method}`);
      this.port?.postMessage({
        type: 'rpc_response',
        id,
        error: {
          code: -32601,
          message: `Method not found: ${method}`,
        },
      });
      return;
    }

    // Call the handler and send the response
    try {
      const request: RpcRequest = { id, method, params };
      const response = await handler(request);

      // Make sure the response includes the request ID
      response.id = id;

      console.log(`[McpHostManager] Send RPC response:`, response);

      this.port?.postMessage({
        type: 'rpc_response',
        ...response,
      });
    } catch (error) {
      console.error(`[McpHostManager] Error handling RPC method ${method}:`, error);
      this.port?.postMessage({
        type: 'rpc_response',
        id,
        error: {
          code: -32603,
          message: error instanceof Error ? error.message : String(error),
        },
      });
    }
  }

  /**
   * Processes an incoming RPC response from the MCP Host.
   * @param data The RPC response data
   */
  private handleRpcResponse(data: unknown): void {
    if (
      typeof data !== "object" ||
      data === null ||
      !("id" in data)
    ) {
      console.warn("[McpHostManager] Invalid RPC response data", data);
      return;
    }
    const { id, error } = data as { id?: string; result?: unknown; error?: unknown };
    console.debug(`[McpHostManager] Received RPC response for ID: ${id}`);

    // Find the pending request for this ID
    const pendingRequest = this.pendingRequests.get(id as string);
    if (!pendingRequest) {
      console.warn(`[McpHostManager] No pending request found for RPC response ID: ${id}`);
      return;
    }

    // Clear the timeout and remove from pending requests
    clearTimeout(pendingRequest.timeoutId);
    this.pendingRequests.delete(id as string);

    // Resolve or reject the promise based on the response
    if (error) {
      pendingRequest.reject(error);
    } else {
      pendingRequest.resolve(data as unknown);
    }
  }

  private handleMessage(message: unknown): void {
    console.debug(`[McpHostManager] Received message:`, message);

    if (typeof message !== "object" || message === null || !("type" in message)) {
      console.log("Unknown message from MCP Host:", message);
      return;
    }

    const msg = message as { type: string; [key: string]: unknown };

    switch (msg.type) {
      case 'status':
        this.updateStatus(msg.data as Partial<McpHostStatus>);
        break;
      case 'error':
        console.error('MCP Host error:', msg.error);
        break;
      // Handle RPC messages
      case 'rpc_request':
        this.handleRpcRequest(msg);
        break;
      case 'rpc_response':
        this.handleRpcResponse(msg);
        break;
      default:
        console.log('Unknown message from MCP Host:', message);
    }
  }

  /**
   * Handles disconnection from the MCP Host.
   * @private
   */
  private handleDisconnect(): void {
    this.stopHeartbeat();
    this.port = null;
    this.updateStatus({
      isConnected: false,
      startTime: null,
    });
  }

  /**
   * Updates the status object and notifies listeners.
   * @param {Partial<McpHostStatus>} updates Status properties to update.
   * @private
   */
  private updateStatus(updates: Partial<McpHostStatus>): void {
    // Update status
    this.status = {
      ...this.status,
      ...updates,
    };

    // Notify listeners
    this.notifyListeners();

    // Broadcast status to popup and other components
    this.broadcastStatus();
  }

  /**
   * Broadcasts status updates to all extension components
   * @private
   */
  private broadcastStatus(): void {
    try {
      chrome.runtime
        .sendMessage({
          type: 'mcpHostStatusUpdate',
          status: this.getStatus(),
        })
        .catch(error => {
          // Ignore errors if no listeners (e.g., popup is closed)
          console.debug('[McpHostManager] No message listeners available:', error.message);
        });
    } catch (error) {
      console.debug('[McpHostManager] Failed to broadcast status:', error);
    }
  }

  /**
   * Notifies all registered listeners of the current status.
   * @private
   */
  private notifyListeners(): void {
    for (const listener of this.listeners) {
      try {
        listener(this.getStatus());
      } catch (error) {
        console.error('Error in status listener:', error);
      }
    }
  }

  /**
   * Starts the heartbeat mechanism.
   * @private
   */
  private startHeartbeat(): void {
    // Clear any existing heartbeat
    this.stopHeartbeat();

    // Send status request
    this.sendStatusRequest();

    // Set up new heartbeat interval
    // Use globalThis to be compatible with both browser and Node.js environments
    this.heartbeatInterval = globalThis.setInterval(() => {
      this.sendStatusRequest();
    }, this.HEARTBEAT_INTERVAL_MS);
  }

  /**
   * Stops the heartbeat mechanism.
   * @private
   */
  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }

    if (this.pingTimeout) {
      clearTimeout(this.pingTimeout);
      this.pingTimeout = null;
    }
  }

  /**
   * Sends a status request to get complete information from MCP Host using RPC.
   * @private
   */
  private sendStatusRequest(): void {
    if (!this.port) {
      this.stopHeartbeat();
      return;
    }

    // Clear any previous ping timeout
    if (this.pingTimeout) {
      clearTimeout(this.pingTimeout);
      this.pingTimeout = null;
    }

    // Use RPC to send status request
    this.rpcRequest(
      {
        method: 'status',
        params: {},
      },
      { timeout: this.PING_TIMEOUT_MS },
    )
      .then(response => {
        console.debug('Status response:', response);

        // Process successful response, update status with complete information
        if (response && typeof response.result === "object" && response.result !== null) {
          const statusData = response.result as { [key: string]: unknown };

          // Convert current_time string to timestamp if needed
          let lastHeartbeat = null;
          if ("current_time" in statusData) {
            const currentTime = statusData.current_time;
            if (typeof currentTime === 'string') {
              lastHeartbeat = new Date(currentTime).getTime();
            } else if (typeof currentTime === 'number') {
              lastHeartbeat = currentTime;
            } else {
              lastHeartbeat = Date.now();
            }
          }

          // Parse start_time from MCP Host response
          let startTime = this.status.startTime; // Keep existing startTime if available
          if ("start_time" in statusData) {
            const startTimeVal = statusData.start_time;
            if (typeof startTimeVal === 'string') {
              startTime = new Date(startTimeVal).getTime();
            } else if (typeof startTimeVal === 'number') {
              startTime = startTimeVal;
            }
          }

          this.updateStatus({
            lastHeartbeat,
            startTime,
            version: typeof statusData.version === "string" ? statusData.version : undefined,
            uptime: typeof statusData.uptime === "string" ? statusData.uptime : undefined,
            ssePort: typeof statusData.sse_port === "string" ? statusData.sse_port : undefined,
            sseBaseURL: typeof statusData.sse_base_url === "string" ? statusData.sse_base_url : undefined,
          });
        }
      })
      .catch(error => {
        // Connection might be lost or request timed out
        console.log('Status request failed:', error.message || 'No response from MCP Host');

        // Mark as disconnected
        this.updateStatus({ isConnected: false });

        // Stop the heartbeat timers
        this.stopHeartbeat();

        // Clean up port reference since we're no longer connected
        this.port = null;
      });
  }
}
