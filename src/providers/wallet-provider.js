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
  
  // Base wallet API implementation
  const createWalletAPI = (chain) => {
    const wallet = {
      isAlgonius: true,
      isPhantom: true,
      isConnected: false,
      autoRefreshOnNetworkChange: false,
      _isConnected: false,
      _publicKey: null,
      _network: null,
      _requestId: 0,
      _callbacks: {},
      _chain: chain, // Track which chain this instance represents

      // Send request to extension with chain information
      _sendRequest: function(method, params) {
        return new Promise((resolve, reject) => {
          const requestId = this._requestId++;
          this._callbacks[requestId] = { resolve, reject };
          
          console.log(`Sending ${this._chain} request to content script:`, { id: requestId, method, params });
          
          // Send message to content script with chain information
          window.postMessage({
            type: "ALGONIUS_WALLET_REQUEST",
            id: requestId,
            method,
            params,
            chain: this._chain // Include chain info in the request
          }, window.location.origin);
        });
      },
      
      // Handle response from extension
      _handleResponse: function(response) {
        console.log(`Received ${this._chain} response from content script:`, response);
        const callback = this._callbacks[response.id];
        if (!callback) {
          console.warn('No callback found for response ID:', response.id);
          return;
        }
        
        delete this._callbacks[response.id];
        
        if (response.error) {
          // If error is a string, create an Error object
          if (typeof response.error === 'string') {
            callback.reject(new Error(response.error));
          } else if (typeof response.error === 'object' && response.error !== null) {
            // If error is an object, try to extract message or stringify it
            const errorMessage = response.error.message || JSON.stringify(response.error);
            callback.reject(new Error(errorMessage));
          } else {
            callback.reject(new Error('Unknown error'));
          }
        } else {
          // Special handling for Solana responses to ensure correct format
          if (this._chain === 'solana' && response.result) {
            // Ensure the result has the correct format for Solana
            let solanaResult = response.result;
            
            // If we have a signature and publicKey, make sure they're in the right format
            if (solanaResult.signature && solanaResult.publicKey) {
              // For Solana, the signature should be a base58 string without 0x prefix
              // The publicKey should also be a base58 string without 0x prefix
              if (typeof solanaResult.signature === 'string' && solanaResult.signature.startsWith('0x')) {
                solanaResult.signature = solanaResult.signature.substring(2); // Remove 0x prefix
              }
              
              if (typeof solanaResult.publicKey === 'string' && solanaResult.publicKey.startsWith('0x')) {
                solanaResult.publicKey = solanaResult.publicKey.substring(2); // Remove 0x prefix
              }
              
              // Ensure signature is a proper 64-byte array for Solana
              // If we have a base58 signature, decode it to bytes for proper format
              try {
                // We need to check if the signature is base58 encoded
                if (typeof solanaResult.signature === 'string') {
                  // This would require a base58 decoding library in the browser
                  // For now, we'll return it as-is but ensure it's properly formatted
                  console.log('Solana signature (base58):', solanaResult.signature);
                }
              } catch (decodeError) {
                console.error('Failed to decode Solana signature:', decodeError);
              }
            }
            
            callback.resolve(solanaResult);
          } else {
            callback.resolve(response.result);
          }
        }
      },

      async request(args) {
        const { method, params } = args;
        console.log(`Algonius Wallet ${this._chain} request called with method:`, method, 'params:', params);
        return this._sendRequest(method, params);
      },

      async connect() {
        console.log(`Algonius Wallet ${this._chain} connect called`);
        try {
          // For different chains, we might need different connection methods
          let accounts;
          if (this._chain === 'solana') {
            accounts = await this.request({ method: 'solana_requestAccounts' });
          } else {
            accounts = await this.request({ method: 'eth_requestAccounts' });
          }
          
          console.log('Received accounts:', accounts);
          if (Array.isArray(accounts) && accounts.length > 0) {
            this._publicKey = accounts[0];
            this._isConnected = true;
            console.log(`${this._chain} wallet connected with public key:`, this._publicKey);
          }
          
          // For Solana, return the public key in the correct format
          if (this._chain === 'solana' && this._publicKey) {
            return { 
              publicKey: this._createSolanaPublicKey(this._publicKey), 
              isConnected: this._isConnected 
            };
          } else {
            return { 
              publicKey: this._publicKey, 
              isConnected: this._isConnected 
            };
          }
        } catch (error) {
          console.error('Failed to connect:', error);
          throw new Error(`Failed to connect: ${error.message}`);
        }
      },

      async disconnect() {
        console.log(`Algonius Wallet ${this._chain} disconnect called`);
        this._publicKey = null;
        this._isConnected = false;
        console.log(`${this._chain} wallet disconnected`);
      },

      // Add Solana-specific connect method for better compatibility
      async connectSolana() {
        if (this._chain === 'solana') {
          console.log('Algonius Wallet Solana connect called');
          return await this.connect();
        } else {
          throw new Error('connectSolana is only available for Solana chain');
        }
      },

      // Signer methods that Phantom provides
      async signMessage(message, encoding = 'utf8') {
        console.log(`Algonius Wallet ${this._chain} signMessage called with message:`, message, 'encoding:', encoding);
        try {
          let result;
          if (this._chain === 'solana') {
            // For Solana, we need to handle the message differently
            // Solana expects a Uint8Array for the message
            let messageBytes;
            if (typeof message === 'string') {
              // If it's a string, convert to Uint8Array
              messageBytes = new TextEncoder().encode(message);
            } else if (message instanceof Uint8Array) {
              // If it's already a Uint8Array, use it directly
              messageBytes = message;
            } else {
              // Otherwise, convert to string then to Uint8Array
              messageBytes = new TextEncoder().encode(String(message));
            }
            
            // Solana signMessage method
            result = await this.request({ 
              method: 'signMessage', 
              params: [Array.from(messageBytes)] // Convert to array for JSON serialization
            });
            
            // For Solana, we need to return the signature in the correct format
            if (result && result.signature && result.publicKey) {
              // GMGN and other Solana dApps expect:
              // 1. signature as Uint8Array (64 bytes)
              // 2. publicKey as base58 string (not with 0x prefix)
              
              // Convert signature from array of integers back to Uint8Array
              let signatureBytes;
              if (Array.isArray(result.signature)) {
                signatureBytes = new Uint8Array(result.signature);
              } else {
                // If it's not an array, we have a problem with our implementation
                throw new Error('Invalid signature format: expected array of integers');
              }
              
              // Ensure the signature is exactly 64 bytes
              if (signatureBytes.length !== 64) {
                throw new Error(`Invalid signature length: expected 64 bytes, got ${signatureBytes.length}`);
              }
              
              // Return the properly formatted result
              const formattedResult = {
                signature: signatureBytes, // Uint8Array(64)
                publicKey: result.publicKey // Base58 string without 0x prefix
              };
              
              console.log(`${this._chain} message signed successfully:`, formattedResult);
              return formattedResult;
            }
          } else {
            // For Ethereum and other EVM chains, use personal_sign
            result = await this.request({ 
              method: 'personal_sign', 
              params: [message, this._publicKey] 
            });
          }
          console.log(`${this._chain} message signed successfully:`, result);
          return result;
        } catch (error) {
          console.error('Failed to sign message:', error);
          throw new Error(`Failed to sign message: ${error.message}`);
        }
      },
      
      async signTransaction(transaction) {
        console.log(`Algonius Wallet ${this._chain} signTransaction called with transaction:`, transaction);
        try {
          let result;
          if (this._chain === 'solana') {
            // For Solana, use signTransaction method
            result = await this.request({ 
              method: 'signTransaction', 
              params: [transaction] 
            });
          } else {
            // For Ethereum and other EVM chains, use eth_signTransaction
            result = await this.request({ 
              method: 'eth_signTransaction', 
              params: [transaction] 
            });
          }
          console.log(`${this._chain} transaction signed successfully:`, result);
          return result;
        } catch (error) {
          console.error('Failed to sign transaction:', error);
          throw new Error(`Failed to sign transaction: ${error.message}`);
        }
      },
      
      async signAllTransactions(transactions) {
        console.log(`Algonius Wallet ${this._chain} signAllTransactions called with transactions:`, transactions);
        try {
          let result;
          if (this._chain === 'solana') {
            // For Solana, use signAllTransactions method
            result = await this.request({ 
              method: 'signAllTransactions', 
              params: [transactions] 
            });
          } else {
            // For Ethereum and other EVM chains, use eth_signTransactions
            result = await this.request({ 
              method: 'eth_signTransactions', 
              params: [transactions] 
            });
          }
          console.log(`${this._chain} all transactions signed successfully:`, result);
          return result;
        } catch (error) {
          console.error('Failed to sign transactions:', error);
          throw new Error(`Failed to sign transactions: ${error.message}`);
        }
      },

      // Add Solana-specific signAndSendTransaction method
      async signAndSendTransaction(transaction, options) {
        console.log(`Algonius Wallet ${this._chain} signAndSendTransaction called with transaction:`, transaction, 'options:', options);
        try {
          if (this._chain === 'solana') {
            // For Solana, use signAndSendTransaction method
            const result = await this.request({ 
              method: 'signAndSendTransaction', 
              params: [transaction, options || {}] 
            });
            console.log(`${this._chain} transaction signed and sent successfully:`, result);
            return result;
          } else {
            throw new Error('signAndSendTransaction is only available for Solana chain');
          }
        } catch (error) {
          console.error('Failed to sign and send transaction:', error);
          throw new Error(`Failed to sign and send transaction: ${error.message}`);
        }
      },

      get publicKey() {
        console.log(`Getting ${this._chain} public key:`, this._publicKey);
        // For Solana, return a PublicKey-like object
        if (this._chain === 'solana' && this._publicKey) {
          return this._createSolanaPublicKey(this._publicKey);
        }
        return this._publicKey;
      },

      get network() {
        console.log(`Getting ${this._chain} network:`, this._network);
        return this._network;
      },
      
      // Add standard properties for better compatibility
      get connected() {
        return this._isConnected;
      },
      
      // Create a Solana PublicKey-like object
      _createSolanaPublicKey: function(publicKeyString) {
        if (this._chain !== 'solana') {
          return publicKeyString;
        }
        
        // Simple base58 alphabet
        const ALPHABET = '123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz';
        
        // Simple base58 decode function (simplified for browser environment)
        function base58ToBytes(str) {
          // This is a simplified implementation
          // In a production environment, you would want to use a proper library
          if (!str || typeof str !== 'string') {
            return new Uint8Array(32); // Return empty 32-byte array
          }
          
          try {
            // For now, we'll create a deterministic byte array from the string
            // This is not a real base58 decode but will satisfy the interface
            const encoder = new TextEncoder();
            const bytes = encoder.encode(str);
            const result = new Uint8Array(32);
            
            // Copy bytes to result array (pad or truncate as needed)
            for (let i = 0; i < Math.min(bytes.length, 32); i++) {
              result[i] = bytes[i];
            }
            
            return result;
          } catch (error) {
            console.warn('Error in base58ToBytes:', error);
            return new Uint8Array(32); // Return empty 32-byte array
          }
        }
        
        // Create an object that mimics Solana's PublicKey interface
        return {
          _bn: publicKeyString, // Store the original string
          toString: function() {
            return publicKeyString;
          },
          toBase58: function() {
            return publicKeyString;
          },
          toBytes: function() {
            // Convert base58 string to bytes
            return base58ToBytes(publicKeyString);
          },
          equals: function(other) {
            return other.toString() === publicKeyString;
          }
        };
      },
      
      // Add standard events for better compatibility
      on: function(event, callback) {
        // Basic event handling for compatibility
        console.log(`Event listener added for ${event} on ${this._chain}`);
        // In a full implementation, we would store and manage event listeners
        // For now, we'll just log the event registration
      },
      
      off: function(event, callback) {
        // Basic event handling for compatibility
        console.log(`Event listener removed for ${event} on ${this._chain}`);
      }
    };
    
    // Add chain-specific properties
    if (chain === 'solana') {
      // Add Solana-specific properties
      wallet.isSolana = true;
      wallet.isSolanaPhantom = true;
    } else if (chain === 'ethereum') {
      // Add Ethereum-specific properties
      wallet.isMetaMask = true;
    }
    
    return wallet;
  };

  // Listen for responses from content script
  window.addEventListener("message", (event) => {
    // Strictly validate message source
    if (event.source !== window) return;
    
    if (event.data && event.data.type === "ALGONIUS_WALLET_RESPONSE") {
      // Route response to appropriate chain instance
      // In a more sophisticated implementation, we might include chain info in the response
      // For now, we'll route to all instances as they share the same callback system
      // This is a simplification - in a production implementation, we'd be more precise
      if (window.algoniusWalletInstances) {
        for (const instance of window.algoniusWalletInstances) {
          instance._handleResponse(event.data);
        }
      }
    }
  });

  // Create separate instances for each chain
  const ethereumWallet = createWalletAPI('ethereum');
  const solanaWallet = createWalletAPI('solana');
  const bitcoinWallet = createWalletAPI('bitcoin');
  const suiWallet = createWalletAPI('sui');
  
  // Store instances for message routing
  window.algoniusWalletInstances = [ethereumWallet, solanaWallet, bitcoinWallet, suiWallet];

  // Expose API to window
  window.algoniusWallet = ethereumWallet; // Default to Ethereum for backward compatibility
  console.log('Algonius Wallet object attached to window');

  // For Phantom compatibility, expose separate objects for each chain
  window.phantom = {
    solana: solanaWallet,
    app: ethereumWallet, // Phantom's app namespace typically refers to Ethereum
    ethereum: ethereumWallet,
    bitcoin: bitcoinWallet,
    sui: suiWallet
  };
  
  // Add additional Solana-specific properties for better Phantom compatibility
  if (solanaWallet) {
    // Add isPhantom property to solana wallet instance
    solanaWallet.isPhantom = true;
    
    // Add standard Solana object to window if it doesn't exist
    if (!window.solana) {
      window.solana = solanaWallet;
    }
  }
  console.log('Phantom compatibility object attached to window');

  // Also expose as window.ethereum for broader compatibility
  window.ethereum = ethereumWallet;
  console.log('Ethereum compatibility object attached to window');

  // Auto-connect for known platforms
  if (window.location.hostname.match(/dexscreener\.com|gmgn\.ai|jup\.ag|uniswap\.org|1inch\.io/)) {
    console.log('Auto-connecting on:', window.location.origin);
    // Use setTimeout to ensure page is fully loaded before attempting connection
    setTimeout(() => {
      // For Solana-based sites like jup.ag, connect the Solana wallet
      if (window.location.hostname.includes('jup.ag')) {
        solanaWallet.connect().catch(error => {
          console.error('Solana auto-connect failed:', error);
          // Fallback to Ethereum wallet if needed
          ethereumWallet.connect().catch(error => {
            console.error('Ethereum auto-connect failed:', error);
          });
        });
      } else {
        // For other sites, connect the Ethereum wallet
        ethereumWallet.connect().catch(error => {
          console.error('Auto-connect failed:', error);
        });
      }
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