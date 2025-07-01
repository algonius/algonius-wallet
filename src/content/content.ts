/**
 * Content Script - Main entry point for wallet injection (TypeScript)
 */

const shouldInject = window.location.hostname.match(
  /dexscreener\.com|gmgn\.ai|jupiter\.ag|uniswap\.org|1inch\.io/,
);

if (shouldInject) {
  // Create script element for wallet provider
  const script = document.createElement('script');
  script.src = chrome.runtime.getURL('wallet-provider.js');
  script.onload = function () {
    (this as HTMLScriptElement).remove();
  };

  // Inject into page
  (document.head || document.documentElement).appendChild(script);

  // Listen for messages from injected script
  window.addEventListener('message', (event: MessageEvent) => {
    if (event.source !== window) return;

    if (
      typeof event.data === 'object' &&
      event.data &&
      'type' in event.data &&
      typeof event.data.type === 'string' &&
      event.data.type.startsWith('ALGONIUS_WALLET_')
    ) {
      chrome.runtime.sendMessage({
        type: 'ALGONIUS_WALLET_FORWARD',
        data: event.data,
      });
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
  });

  console.log('Algonius Wallet content script loaded');
}
