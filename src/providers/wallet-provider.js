/**
 * Wallet API - Injected into page context to provide wallet functionality
 * This script runs in the page's context, not the extension's content script context
 */

(() => {
  // Check if already injected
  if (window.algoniusWalletInjected) {
    console.log('Algonius Wallet already injected, skipping');
    return;
  }
  window.algoniusWalletInjected = true;
  console.log('Algonius Wallet injection starting...');

  console.log("This is running in page context:", window === window.top);
  
  // Wallet API implementation
  const algoniusWallet = {
    isAlgonius: true,
    isPhantom: true,
    isConnected: false,
    autoRefreshOnNetworkChange: false,
    _isConnected: false,
    _publicKey: null,
    _network: null,
    _requestId: 0,
    _callbacks: {},

    // Send request to extension
    _sendRequest: function(method, params) {
      return new Promise((resolve, reject) => {
        const requestId = this._requestId++;
        this._callbacks[requestId] = { resolve, reject };
        
        console.log('Sending request to content script:', { id: requestId, method, params });
        
        // Send message to content script
        window.postMessage({
          type: "ALGONIUS_WALLET_REQUEST",
          id: requestId,
          method,
          params
        }, window.location.origin);
      });
    },
    
    // Handle response from extension
    _handleResponse: function(response) {
      console.log('Received response from content script:', response);
      const callback = this._callbacks[response.id];
      if (!callback) {
        console.warn('No callback found for response ID:', response.id);
        return;
      }
      
      delete this._callbacks[response.id];
      
      if (response.error) {
        callback.reject(new Error(response.error));
      } else {
        callback.resolve(response.result);
      }
    },

    async request(args) {
      const { method, params } = args;
      console.log('Algonius Wallet request called with method:', method, 'params:', params);
      return this._sendRequest(method, params);
    },

    async connect() {
      console.log('Algonius Wallet connect called');
      try {
        const accounts = await this.request({ method: 'eth_requestAccounts' });
        console.log('Received accounts:', accounts);
        if (Array.isArray(accounts) && accounts.length > 0) {
          this._publicKey = accounts[0];
          this._isConnected = true;
          console.log('Wallet connected with public key:', this._publicKey);
        }
        return { 
          publicKey: this._publicKey, 
          isConnected: this._isConnected 
        };
      } catch (error) {
        console.error('Failed to connect:', error);
        throw new Error(`Failed to connect: ${error.message}`);
      }
    },

    async disconnect() {
      console.log('Algonius Wallet disconnect called');
      this._publicKey = null;
      this._isConnected = false;
      console.log('Wallet disconnected');
    },

    // Signer methods that Phantom provides
    async signMessage(message, encoding = 'utf8') {
      console.log('Algonius Wallet signMessage called with message:', message, 'encoding:', encoding);
      try {
        const result = await this.request({ 
          method: 'personal_sign', 
          params: [message, this._publicKey] 
        });
        console.log('Message signed successfully:', result);
        return result;
      } catch (error) {
        console.error('Failed to sign message:', error);
        throw new Error(`Failed to sign message: ${error.message}`);
      }
    },
    
    async signTransaction(transaction) {
      console.log('Algonius Wallet signTransaction called with transaction:', transaction);
      try {
        const result = await this.request({ 
          method: 'eth_signTransaction', 
          params: [transaction] 
        });
        console.log('Transaction signed successfully:', result);
        return result;
      } catch (error) {
        console.error('Failed to sign transaction:', error);
        throw new Error(`Failed to sign transaction: ${error.message}`);
      }
    },
    
    async signAllTransactions(transactions) {
      console.log('Algonius Wallet signAllTransactions called with transactions:', transactions);
      try {
        const result = await this.request({ 
          method: 'eth_signTransactions', 
          params: [transactions] 
        });
        console.log('All transactions signed successfully:', result);
        return result;
      } catch (error) {
        console.error('Failed to sign transactions:', error);
        throw new Error(`Failed to sign transactions: ${error.message}`);
      }
    },

    get publicKey() {
      console.log('Getting public key:', this._publicKey);
      return this._publicKey;
    },

    get network() {
      console.log('Getting network:', this._network);
      return this._network;
    },
  };

  // Listen for responses from content script
  window.addEventListener("message", (event) => {
    // Strictly validate message source
    if (event.source !== window) return;
    
    if (event.data && event.data.type === "ALGONIUS_WALLET_RESPONSE") {
      algoniusWallet._handleResponse(event.data);
    }
  });

  // Expose API to window
  window.algoniusWallet = algoniusWallet;
  console.log('Algonius Wallet object attached to window');

  // For Phantom compatibility, also expose as window.phantom
  window.phantom = {
    solana: algoniusWallet,
    app: algoniusWallet,
    ethereum: algoniusWallet,
    bitcoin: algoniusWallet,
    sui: algoniusWallet
  };
  console.log('Phantom compatibility object attached to window');

  // Also expose as window.ethereum for broader compatibility
  window.ethereum = algoniusWallet;
  console.log('Ethereum compatibility object attached to window');

  // Auto-connect for known platforms
  if (window.location.hostname.match(/dexscreener\.com|gmgn\.ai|jupiter\.ag|uniswap\.org|1inch\.io/)) {
    console.log('Auto-connecting on:', window.location.origin);
    // Use setTimeout to ensure page is fully loaded before attempting connection
    setTimeout(() => {
      algoniusWallet.connect().catch(error => {
        console.error('Auto-connect failed:', error);
      });
    }, 1000);
  }

  // Debug logging
  console.log('Algonius Wallet API injected into page context');
  console.log('Algonius Wallet object:', window.algoniusWallet);
  console.log('Phantom object:', window.phantom);
  console.log('Ethereum object:', window.ethereum);
  
  // Additional check for Phantom detection
  console.log('Phantom detection - isPhantom:', window.phantom?.ethereum?.isPhantom);
  console.log('Ethereum detection - isPhantom:', window.ethereum?.isPhantom);
})();