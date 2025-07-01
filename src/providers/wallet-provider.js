/**
 * Wallet Provider - Injected into DeFi platforms
 * Provides Web3-compatible interface for Algonius Wallet
 */

class AlgoniusWalletProvider {
  constructor() {
    this.accounts = [];
    this.chainId = null;
    this.isConnected = false;
    this.providerName = 'AlgoniusWallet';
    this.version = '1.0.0';

    // Setup message listeners
    this.setupListeners();
  }

  setupListeners() {
    window.addEventListener('message', (event) => {
      if (event.source !== window) return;

      if (event.data.type && event.data.type.startsWith('ALGONIUS_WALLET_')) {
        this.handleMessage(event.data);
      }
    });
  }

  async handleMessage(message) {
    try {
      const { type, id, data } = message;

      switch (type) {
        case 'ALGONIUS_WALLET_REQUEST':
          const response = await this.handleRequest(data.method, data.params);
          window.postMessage(
            {
              type: 'ALGONIUS_WALLET_RESPONSE',
              id,
              data: response,
            },
            '*',
          );
          break;

        case 'ALGONIUS_WALLET_EVENT':
          this.emit(data.event, data.payload);
          break;
      }
    } catch (error) {
      console.error('Algonius Wallet Provider error:', error);
      window.postMessage(
        {
          type: 'ALGONIUS_WALLET_ERROR',
          id: message.id,
          error: error.message,
        },
        '*',
      );
    }
  }

  async handleRequest(method, params) {
    // Forward request to background script
    const response = await chrome.runtime.sendMessage({
      type: 'MCP_COMMAND',
      data: {
        command: 'provider.' + method,
        params,
      },
    });

    if (response.success) {
      return response.data;
    } else {
      throw new Error(response.error || 'Request failed');
    }
  }

  // Web3 Provider Interface
  async request(args) {
    return this.handleRequest('request', args);
  }

  async enable() {
    const accounts = await this.request({ method: 'eth_requestAccounts' });
    this.accounts = accounts;
    this.isConnected = true;
    return accounts;
  }

  async send(method, params) {
    return this.request({ method, params });
  }

  async sendAsync(payload, callback) {
    try {
      const result = await this.request(payload);
      callback(null, result);
    } catch (error) {
      callback(error);
    }
  }

  // Event emitter
  on(event, handler) {
    if (!this._events) this._events = {};
    if (!this._events[event]) this._events[event] = [];
    this._events[event].push(handler);
  }

  off(event, handler) {
    if (!this._events || !this._events[event]) return;
    this._events[event] = this._events[event].filter((h) => h !== handler);
  }

  emit(event, payload) {
    if (!this._events || !this._events[event]) return;
    this._events[event].forEach((handler) => handler(payload));
  }
}

// Inject provider into page
if (typeof window.algoniusWallet === 'undefined') {
  window.algoniusWallet = new AlgoniusWalletProvider();

  // Also expose as EIP-1193 provider
  if (typeof window.ethereum === 'undefined') {
    window.ethereum = window.algoniusWallet;
  }
}
