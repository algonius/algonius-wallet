/**
 * Wallet Provider - Injected into DeFi platforms
 * Provides Web3-compatible interface for Algonius Wallet
 * Simulates Phantom wallet for compatibility with sites like gmgn.ai
 */

class AlgoniusWalletProvider {
  constructor() {
    this.accounts = [];
    this.chainId = null;
    this.isConnected = false;
    this.providerName = 'phantom';
    this.version = '1.0.0';
    this.isPhantom = true; // Mark as Phantom-compatible
    this.autoRefreshOnNetworkChange = false;
    this._events = {};
    
    // Additional Phantom properties
    this._isConnected = false;
    this._publicKey = null;
    this._network = null;

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
      // Update local state based on method
      if (method === 'eth_requestAccounts' || method === 'eth_accounts') {
        if (response.data && Array.isArray(response.data) && response.data.length > 0) {
          this._isConnected = true;
          this._publicKey = response.data[0];
        }
      }
      
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
    try {
      const accounts = await this.request({ method: 'eth_requestAccounts' });
      this.accounts = accounts;
      this.isConnected = true;
      this._isConnected = true;
      if (accounts && accounts.length > 0) {
        this._publicKey = accounts[0];
      }
      this.emit('connect', { chainId: this.chainId });
      return accounts;
    } catch (error) {
      throw new Error(`Failed to enable wallet: ${error.message}`);
    }
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
    
    // For Phantom compatibility, also emit connect event when accounts are available
    if (event === 'accountsChanged' && this.accounts.length > 0) {
      setTimeout(() => handler(this.accounts), 0);
    }
  }

  off(event, handler) {
    if (!this._events || !this._events[event]) return;
    this._events[event] = this._events[event].filter((h) => h !== handler);
  }

  emit(event, payload) {
    if (!this._events || !this._events[event]) return;
    this._events[event].forEach((handler) => handler(payload));
    
    // For Phantom compatibility, also emit connect event
    if (event === 'connect') {
      this.isConnected = true;
      this._isConnected = true;
    } else if (event === 'disconnect') {
      this.isConnected = false;
      this._isConnected = false;
      this.accounts = [];
      this._publicKey = null;
    } else if (event === 'accountsChanged') {
      this.accounts = payload || [];
      if (this.accounts && this.accounts.length > 0) {
        this._publicKey = this.accounts[0];
        this._isConnected = true;
      } else {
        this._publicKey = null;
        this._isConnected = false;
      }
    }
  }
  
  // Phantom-specific methods and properties
  async connect() {
    try {
      const accounts = await this.request({ method: 'eth_requestAccounts' });
      this.accounts = accounts;
      this.isConnected = true;
      this._isConnected = true;
      if (accounts && accounts.length > 0) {
        this._publicKey = accounts[0];
      }
      this.emit('connect', { chainId: this.chainId });
      return { 
        publicKey: this._publicKey, 
        isConnected: this._isConnected 
      };
    } catch (error) {
      throw new Error(`Failed to connect: ${error.message}`);
    }
  }
  
  async disconnect() {
    // Note: This is a simplified implementation
    // A full implementation would need to communicate with the extension to actually disconnect
    this.accounts = [];
    this.isConnected = false;
    this._isConnected = false;
    this._publicKey = null;
    this.emit('disconnect');
  }
  
  // Additional Phantom methods
  async isConnectedAndUnlocked() {
    try {
      // Try to get accounts without user interaction
      await this.request({ method: 'eth_accounts' });
      return this.accounts.length > 0;
    } catch (error) {
      return false;
    }
  }
  
  // Getter for Phantom compatibility
  get publicKey() {
    return this._publicKey;
  }
  
  get network() {
    return this._network;
  }
  
  // Signer methods that Phantom provides
  async signMessage(message, encoding = 'utf8') {
    // Implementation would go here
    throw new Error('signMessage not implemented');
  }
  
  async signTransaction(transaction) {
    // Implementation would go here
    throw new Error('signTransaction not implemented');
  }
  
  async signAllTransactions(transactions) {
    // Implementation would go here
    throw new Error('signAllTransactions not implemented');
  }
}

// Inject provider into page
if (typeof window.algoniusWallet === 'undefined') {
  window.algoniusWallet = new AlgoniusWalletProvider();

  // Also expose as EIP-1193 provider
  if (typeof window.ethereum === 'undefined') {
    window.ethereum = window.algoniusWallet;
  }
  
  // For Phantom compatibility, also expose as window.phantom
  if (typeof window.phantom === 'undefined') {
    window.phantom = {
      ethereum: window.algoniusWallet
    };
  }
}
