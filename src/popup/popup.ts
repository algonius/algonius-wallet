/**
 * Popup Script - Handles UI interactions for Algonius Wallet (TypeScript)
 */

interface Balance {
  token: string;
  amount: string;
}

interface WalletState {
  balances: Balance[];
}

interface Signal {
  strategy: string;
  status: string;
  tokenIn: string;
  tokenOut: string;
  amount: string;
}

interface ConnectionStatus {
  connected: boolean;
  network?: string;
}

document.addEventListener('DOMContentLoaded', async () => {
  // Initialize UI
  const refreshBtn = document.getElementById('refresh-btn') as HTMLButtonElement | null;
  const settingsBtn = document.getElementById('settings-btn') as HTMLButtonElement | null;
  const balanceDisplay = document.getElementById('balance-display') as HTMLElement | null;
  const signalsList = document.getElementById('signals-list') as HTMLElement | null;
  const lastUpdated = document.getElementById('last-updated') as HTMLElement | null;
  const connectionStatus = document.getElementById('connection-status') as HTMLElement | null;

  // Connect to background script
  const port = chrome.runtime.connect({ name: 'popup' });

  // Handle messages from background
  port.onMessage.addListener((message: unknown) => {
    if (
      typeof message === 'object' &&
      message !== null &&
      'type' in message
    ) {
      switch ((message as { type: string }).type) {
        case 'WALLET_STATE':
          updateWalletState((message as unknown as { data: WalletState }).data);
          break;
        case 'SIGNALS_UPDATE':
          updateSignalsList((message as unknown as { data: Signal[] }).data);
          break;
        case 'CONNECTION_STATUS':
          updateConnectionStatus((message as unknown as { data: ConnectionStatus }).data);
          break;
      }
    }
  });

  // Request initial data
  refreshData();

  // Button event listeners
  refreshBtn?.addEventListener('click', refreshData);
  settingsBtn?.addEventListener('click', openSettings);

  // Functions
  async function refreshData() {
    try {
      // Get wallet state
      const walletState: { data: WalletState } = await chrome.runtime.sendMessage({
        type: 'GET_WALLET_STATE',
      });
      updateWalletState(walletState.data);

      // Get signals
      const signals: { data: Signal[] } = await chrome.runtime.sendMessage({
        type: 'GET_SIGNALS',
      });
      updateSignalsList(signals.data);

      // Update timestamp
      if (lastUpdated) {
        lastUpdated.textContent = `Last updated: ${new Date().toLocaleTimeString()}`;
      }
    } catch (error) {
      console.error('Error refreshing data:', error);
    }
  }

  function updateWalletState(state: WalletState) {
    if (!balanceDisplay) return;
    balanceDisplay.innerHTML = '';

    if (state.balances && state.balances.length > 0) {
      state.balances.forEach((balance) => {
        const balanceEl = document.createElement('div');
        balanceEl.className = 'balance-item';
        balanceEl.innerHTML = `
          <span class="token-name">${balance.token}</span>
          <span class="token-amount">${balance.amount}</span>
        `;
        balanceDisplay.appendChild(balanceEl);
      });
    } else {
      balanceDisplay.innerHTML = '<div class="empty-state">No balances found</div>';
    }
  }

  function updateSignalsList(signals: Signal[]) {
    if (!signalsList) return;
    signalsList.innerHTML = '';

    if (signals && signals.length > 0) {
      signals.forEach((signal) => {
        const signalEl = document.createElement('div');
        signalEl.className = 'signal-item';
        signalEl.innerHTML = `
          <div class="signal-header">
            <span class="signal-strategy">${signal.strategy}</span>
            <span class="signal-status ${signal.status}">${signal.status}</span>
          </div>
          <div class="signal-details">
            <span>${signal.tokenIn} → ${signal.tokenOut}</span>
            <span>${signal.amount}</span>
          </div>
        `;
        signalsList.appendChild(signalEl);
      });
    } else {
      signalsList.innerHTML = '<div class="empty-state">No active signals</div>';
    }
  }

  function updateConnectionStatus(status: ConnectionStatus) {
    if (!connectionStatus) return;
    connectionStatus.className = `status-dot ${status.connected ? 'connected' : 'disconnected'}`;
    const networkName = document.getElementById('network-name');
    if (networkName) {
      networkName.textContent = status.network || 'Disconnected';
    }
  }

  function openSettings() {
    chrome.runtime.openOptionsPage();
  }
});
