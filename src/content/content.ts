/**
 * Content Script - Main entry point for wallet injection (TypeScript)
 */

// Check if we should inject based on hostname or for development/debugging
const shouldInject = window.location.hostname.match(
  /dexscreener\.com|gmgn\.ai|jupiter\.ag|uniswap\.org|1inch\.io/
) || process.env.NODE_ENV === 'development';

// Always inject for now to support Phantom compatibility testing
// In production, we might want to be more selective
if (shouldInject) {
  console.log('Algonius Wallet injecting wallet API into page');
  
  // Create script element for wallet API
  const script = document.createElement('script');
  script.src = chrome.runtime.getURL('providers/wallet-provider.js');
  script.onload = function () {
    console.log('Algonius Wallet API injected successfully');
    (this as HTMLScriptElement).remove();
  };
  script.onerror = function (error) {
    console.error('Algonius Wallet failed to inject API:', error);
  };

  // Inject into page
  (document.head || document.documentElement).appendChild(script);

  // Listen for messages from the page
  window.addEventListener("message", (event) => {
    // Strictly validate message source
    if (event.source !== window) return;
    
    if (event.data && event.data.type === "ALGONIUS_WALLET_REQUEST") {
      const request = event.data;
      console.log('Content script received request from page:', request);
      
      // Forward the request to the background
      chrome.runtime.sendMessage(
        { 
          type: "ALGONIUS_WALLET_FORWARD", 
          data: request 
        }, 
        (response) => {
          console.log('Content script received response from background:', response);
          // Send the response back to the page
          window.postMessage({
            type: "ALGONIUS_WALLET_RESPONSE",
            id: request.id,
            result: response?.result,
            error: response?.error
          }, window.location.origin);
        }
      );
    }
  });

  console.log('Algonius Wallet content script loaded');
}
