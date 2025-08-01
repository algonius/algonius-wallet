/**
 * Injected Script - Bridges between web pages and Algonius Wallet (TypeScript)
 */

declare global {
  interface Window {
    algoniusWalletInjected?: boolean;
    algoniusWallet?: {
      request: (args: unknown) => Promise<unknown>;
      on: (event: string, handler: (payload: unknown) => void) => () => void;
      removeListener: (event: string, handler: (payload: unknown) => void) => void;
      isPhantom?: boolean;
      isConnected?: boolean;
      autoRefreshOnNetworkChange?: boolean;
      isMetaMask?: boolean;
      enable?: () => Promise<unknown>;
      _events?: Record<string, unknown[]>;
      _publicKey?: string | null;
      _isConnected?: boolean;
      _network?: string | null;
      connect?: () => Promise<unknown>;
      disconnect?: () => Promise<void>;
      signMessage?: (message: unknown, encoding?: string) => Promise<unknown>;
      publicKey?: string | null;
      network?: string | null;
    };
    ethereum?: Window['algoniusWallet'];
    phantom?: {
      ethereum?: Window['algoniusWallet'];
    };
  }
}

(() => {
  // Check if already injected
  if (window.algoniusWalletInjected) return;
  window.algoniusWalletInjected = true;

  // Setup message bridge
  const eventListeners: Map<string, Set<(payload: unknown) => void>> = new Map();

  // Forward messages from page to extension
  window.addEventListener('message', (event: MessageEvent) => {
    if (event.source !== window) return;

    if (event.data && event.data.type === 'ALGONIUS_WALLET_REQUEST') {
      chrome.runtime.sendMessage(
        {
          type: 'ALGONIUS_WALLET_FORWARD',
          data: event.data,
        },
        (response) => {
          window.postMessage(
            {
              type: 'ALGONIUS_WALLET_RESPONSE',
              id: event.data.id,
              data: response,
            },
            '*',
          );
        },
      );
    }
  });

  // Forward messages from extension to page
  chrome.runtime.onMessage.addListener((message: unknown) => {
    if (
      typeof message === 'object' &&
      message !== null &&
      'type' in message &&
      (message as { type: unknown }).type === 'ALGONIUS_WALLET_EVENT'
    ) {
      window.postMessage(message, '*');
    }
    return true;
  });

  // Detect if page is using web3
  function detectWeb3() {
    if (typeof window.ethereum !== 'undefined') {
      window.postMessage(
        {
          type: 'ALGONIUS_WALLET_DETECTED',
          data: {
            isWeb3Available: true,
            provider: 'AlgoniusWallet',
          },
        },
        '*',
      );
    }
  }

  // Initial setup
  detectWeb3();

  // Expose API for direct access
  window.algoniusWallet = {
    async request(args: unknown) {
      return new Promise((resolve, reject) => {
        const id = Date.now().toString();

        const listener = (event: MessageEvent) => {
          if (
            event.data &&
            typeof event.data === 'object' &&
            'type' in event.data &&
            event.data.type === 'ALGONIUS_WALLET_RESPONSE' &&
            event.data.id === id
          ) {
            window.removeEventListener('message', listener as EventListener);
            if ('error' in event.data && event.data.error) {
              reject(new Error(event.data.error));
            } else {
              resolve((event.data as { data?: unknown }).data);
            }
          }
        };

        window.addEventListener('message', listener as EventListener);

        window.postMessage(
          {
            type: 'ALGONIUS_WALLET_REQUEST',
            id,
            data: args,
          },
          '*',
        );
      });
    },

    on(event: string, handler: (payload: unknown) => void) {
      if (!eventListeners.has(event)) {
        eventListeners.set(event, new Set());
      }
      eventListeners.get(event)!.add(handler);

      const listener = (e: MessageEvent) => {
        if (
          e.data &&
          typeof e.data === 'object' &&
          'type' in e.data &&
          e.data.type === 'ALGONIUS_WALLET_EVENT' &&
          e.data.data &&
          typeof e.data.data === 'object' &&
          'event' in e.data.data &&
          e.data.data.event === event
        ) {
          handler((e.data.data as { payload: unknown }).payload);
        }
      };

      window.addEventListener('message', listener as EventListener);
      return () => {
        window.removeEventListener('message', listener as EventListener);
        eventListeners.get(event)!.delete(handler);
      };
    },

    removeListener(event: string, handler: (payload: unknown) => void) {
      if (eventListeners.has(event)) {
        eventListeners.get(event)!.delete(handler);
      }
    },
    
    // Add Phantom-specific properties
    isPhantom: true,
    isConnected: false,
    autoRefreshOnNetworkChange: false,
    _isConnected: false,
    _publicKey: null,
    _network: null,
    
    // Add Phantom-specific methods
    async connect() {
      try {
        const accounts = await this.request({ method: 'eth_requestAccounts' });
        if (Array.isArray(accounts) && accounts.length > 0) {
          this._publicKey = accounts[0];
          this._isConnected = true;
        }
        return { 
          publicKey: this._publicKey, 
          isConnected: this._isConnected 
        };
      } catch (error) {
        throw new Error(`Failed to connect: ${(error as Error).message}`);
      }
    },
    
    async disconnect() {
      this._publicKey = null;
      this._isConnected = false;
    },
    
    // Getter for Phantom compatibility
    get publicKey() {
      return this._publicKey;
    },
    
    get network() {
      return this._network;
    },
  } as Window['algoniusWallet'];

  // For Phantom compatibility, also expose as window.phantom
  if (typeof window.phantom === 'undefined') {
    window.phantom = {
      ethereum: window.algoniusWallet
    };
  }
  
  // Also expose as window.ethereum for broader compatibility
  if (typeof window.ethereum === 'undefined') {
    window.ethereum = window.algoniusWallet;
  }

  // Auto-enable for known DeFi platforms
  if (
    window.location.hostname.match(/dexscreener\.com|jupiter\.ag|uniswap\.org|1inch\.io|gmgn\.ai/)
  ) {
    window.algoniusWallet?.request({ method: 'eth_requestAccounts' }).catch(console.error);
  }
  
  // Debug logging for Phantom detection
  console.log('Algonius Wallet injected as Phantom:', window.phantom);
  console.log('Phantom ethereum provider:', window.phantom?.ethereum);
  console.log('Is Phantom detected:', window.phantom?.ethereum?.isPhantom);
  
  // Log successful injection
  console.log('Algonius Wallet successfully injected and ready to use');
})();
export {};
