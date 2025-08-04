/**
 * Transaction Confirmation Overlay for AI Agent visual feedback
 * Implements REQ-EXT-009 to REQ-EXT-012
 */

import type { PendingTransaction } from '../shared/types';

// Re-export for backward compatibility
export type { PendingTransaction };

/**
 * TransactionOverlay manages the display of pending transaction notifications
 * for AI Agent interaction as specified in REQ-EXT-009 to REQ-EXT-012
 */
export class TransactionOverlay {
  private overlay: HTMLElement | null = null;

  /**
   * Shows pending transaction overlay in bottom-right corner of DApp page
   * REQ-EXT-009: Display transaction confirmation overlay in bottom-right corner
   * REQ-EXT-010: Provide visual prompt for AI Agent to use get_pending_transactions MCP tool
   * REQ-EXT-011: Display transaction details (amount, token, destination)
   */
  showPendingTransaction(transaction: PendingTransaction): void {
    // Remove existing overlay immediately (without animation) if present
    if (this.overlay) {
      this.overlay.remove();
      this.overlay = null;
    }

    // Create overlay element
    this.overlay = document.createElement('div');
    this.overlay.className = 'algonius-transaction-overlay';
    
    // Apply styles for bottom-right positioning and AI Agent branding
    this.overlay.style.cssText = `
      position: fixed;
      bottom: 20px;
      right: 20px;
      width: 300px;
      max-width: calc(100vw - 40px);
      background: #1a1a1a;
      border: 2px solid #00ff88;
      border-radius: 8px;
      padding: 16px;
      color: white;
      font-family: monospace;
      font-size: 12px;
      line-height: 1.4;
      z-index: 10000;
      box-shadow: 0 4px 20px rgba(0, 255, 136, 0.3);
      transition: opacity 0.3s ease-in-out, transform 0.3s ease-in-out;
      transform: translateY(0);
      opacity: 1;
      pointer-events: auto;
      user-select: none;
      -webkit-user-select: none;
      -moz-user-select: none;
      -ms-user-select: none;
    `;
    
    // Add ARIA attributes for accessibility
    this.overlay.setAttribute('role', 'alert');
    this.overlay.setAttribute('aria-live', 'assertive');
    this.overlay.setAttribute('aria-label', 'AI Agent Transaction Pending Notification');

    // Create overlay content with transaction details using semantic HTML
    this.overlay.innerHTML = `
      <header style="font-size: 14px; font-weight: bold; margin-bottom: 8px; display: flex; align-items: center; gap: 8px;">
        <span role="img" aria-label="Robot">🤖</span>
        <span>AI Agent: Transaction Pending</span>
      </header>
      <main style="margin-bottom: 8px;">
        <div style="font-size: 12px; margin-bottom: 4px;" aria-label="Transaction amount">
          <strong>Amount:</strong> ${transaction.amount} ${transaction.token}
        </div>
        <div style="font-size: 12px; margin-bottom: 4px; word-break: break-all;" aria-label="Destination address">
          <strong>To:</strong> ${transaction.to}
        </div>
      </main>
      <footer style="font-size: 10px; color: #888; margin-top: 8px; padding-top: 8px; border-top: 1px solid #333;">
        <em>Use get_pending_transactions MCP tool to review and approve</em>
      </footer>
    `;

    // Append to document body
    document.body.appendChild(this.overlay);
  }

  /**
   * Hides and removes the transaction overlay with smooth animation
   * REQ-EXT-012: Update or remove overlay when AI Agent completes decision
   */
  hideOverlay(): void {
    if (this.overlay) {
      // Animate out before removing
      this.overlay.style.opacity = '0';
      this.overlay.style.transform = 'translateY(20px)';
      
      // Remove after animation completes
      setTimeout(() => {
        if (this.overlay) {
          this.overlay.remove();
          this.overlay = null;
        }
      }, 300); // Match transition duration
    }
  }
}