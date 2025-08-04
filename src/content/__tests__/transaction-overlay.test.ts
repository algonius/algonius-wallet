import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { TransactionOverlay } from '../transaction-overlay';

// Define PendingTransaction interface based on Go struct
interface PendingTransaction {
  hash: string;
  chain: string;
  from: string;
  to: string;
  amount: string;
  token: string;
  type: string;
  status: string;
  confirmations: number;
  required_confirmations: number;
  block_number?: number;
  nonce?: number;
  gas_fee: string;
  priority: string;
  estimated_confirmation_time: string;
  submitted_at: string;
  last_checked: string;
  rejected_at?: string;
  rejection_reason?: string;
  rejection_details?: string;
  rejection_audit_log_id?: string;
}

describe('TransactionOverlay', () => {
  let overlay: TransactionOverlay;
  let mockTransaction: PendingTransaction;

  beforeEach(() => {
    // Setup DOM
    document.body.innerHTML = '';
    
    // Create overlay instance
    overlay = new TransactionOverlay();

    // Mock transaction data
    mockTransaction = {
      hash: '0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456',
      chain: 'ethereum',
      from: '0x742d35Cc6634C0532925a3b8D4C2B79C2b86A7A8',
      to: '0x8ba1f109551bD432803012645Hac136c22C4F9B',
      amount: '0.5',
      token: 'ETH',
      type: 'transfer',
      status: 'pending',
      confirmations: 2,
      required_confirmations: 6,
      gas_fee: '0.0021',
      priority: 'medium',
      estimated_confirmation_time: '2-3 minutes',
      submitted_at: new Date().toISOString(),
      last_checked: new Date().toISOString(),
    };
  });

  afterEach(() => {
    // Clean up overlay
    overlay.hideOverlay();
    document.body.innerHTML = '';
  });

  describe('Class instantiation and DOM manipulation', () => {
    it('should create TransactionOverlay instance', () => {
      expect(overlay).toBeInstanceOf(TransactionOverlay);
    });

    it('should initially have no overlay element', () => {
      const overlayElement = document.querySelector('.algonius-transaction-overlay');
      expect(overlayElement).toBeNull();
    });

    it('should create overlay element when showPendingTransaction is called', () => {
      overlay.showPendingTransaction(mockTransaction);
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay');
      expect(overlayElement).toBeInstanceOf(HTMLElement);
    });

    it('should remove overlay element when hideOverlay is called', async () => {
      overlay.showPendingTransaction(mockTransaction);
      overlay.hideOverlay();
      
      // Wait for animation to complete
      await new Promise(resolve => setTimeout(resolve, 350));
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay');
      expect(overlayElement).toBeNull();
    });

    it('should append overlay to document body', () => {
      overlay.showPendingTransaction(mockTransaction);
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay');
      expect(overlayElement?.parentElement).toBe(document.body);
    });
  });

  describe('Overlay positioning and styling', () => {
    it('should position overlay in bottom-right corner', () => {
      overlay.showPendingTransaction(mockTransaction);
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay') as HTMLElement;
      const computedStyle = window.getComputedStyle(overlayElement);
      
      expect(computedStyle.position).toBe('fixed');
      expect(computedStyle.bottom).toBe('20px');
      expect(computedStyle.right).toBe('20px');
    });

    it('should have proper dimensions and styling', () => {
      overlay.showPendingTransaction(mockTransaction);
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay') as HTMLElement;
      const computedStyle = window.getComputedStyle(overlayElement);
      
      expect(computedStyle.width).toBe('300px');
      expect(computedStyle.backgroundColor).toBe('rgb(26, 26, 26)'); // Browser converts hex to RGB
      expect(computedStyle.border).toBe('2px solid rgb(0, 255, 136)'); // Browser converts hex to RGB
      expect(computedStyle.borderRadius).toBe('8px');
      expect(computedStyle.padding).toBe('16px');
      expect(computedStyle.color).toBe('rgb(255, 255, 255)'); // Browser converts 'white' to RGB
      expect(computedStyle.fontFamily).toBe('monospace');
    });

    it('should have high z-index for overlay positioning', () => {
      overlay.showPendingTransaction(mockTransaction);
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay') as HTMLElement;
      const computedStyle = window.getComputedStyle(overlayElement);
      
      expect(computedStyle.zIndex).toBe('10000');
    });

    it('should have proper box shadow for visual prominence', () => {
      overlay.showPendingTransaction(mockTransaction);
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay') as HTMLElement;
      const computedStyle = window.getComputedStyle(overlayElement);
      
      // Check that box shadow exists and contains the expected values
      expect(computedStyle.boxShadow).toContain('0 4px 20px');
      expect(computedStyle.boxShadow).toContain('rgba(0, 255, 136');
    });
  });

  describe('Transaction details display', () => {
    it('should display AI Agent header message', () => {
      overlay.showPendingTransaction(mockTransaction);
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay') as HTMLElement;
      expect(overlayElement.innerHTML).toContain('AI Agent: Transaction Pending');
      expect(overlayElement.innerHTML).toContain('🤖');
    });

    it('should display transaction amount and token', () => {
      overlay.showPendingTransaction(mockTransaction);
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay') as HTMLElement;
      expect(overlayElement.innerHTML).toContain(`<strong>Amount:</strong> ${mockTransaction.amount} ${mockTransaction.token}`);
    });

    it('should display transaction destination address', () => {
      overlay.showPendingTransaction(mockTransaction);
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay') as HTMLElement;
      expect(overlayElement.innerHTML).toContain(`<strong>To:</strong> ${mockTransaction.to}`);
    });

    it('should display MCP tool instruction message', () => {
      overlay.showPendingTransaction(mockTransaction);
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay') as HTMLElement;
      expect(overlayElement.innerHTML).toContain('Use get_pending_transactions MCP tool to review and approve');
    });

    it('should handle different token types correctly', () => {
      const usdtTransaction = {
        ...mockTransaction,
        amount: '1000',
        token: 'USDT',
        chain: 'bsc'
      };
      
      overlay.showPendingTransaction(usdtTransaction);
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay') as HTMLElement;
      expect(overlayElement.innerHTML).toContain('<strong>Amount:</strong> 1000 USDT');
    });

    it('should handle long addresses by displaying them fully', () => {
      const longAddressTransaction = {
        ...mockTransaction,
        to: '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef'
      };
      
      overlay.showPendingTransaction(longAddressTransaction);
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay') as HTMLElement;
      expect(overlayElement.innerHTML).toContain(`<strong>To:</strong> ${longAddressTransaction.to}`);
    });
  });

  describe('Show/hide functionality', () => {
    it('should handle multiple calls to showPendingTransaction', async () => {
      overlay.showPendingTransaction(mockTransaction);
      const firstOverlay = document.querySelector('.algonius-transaction-overlay');
      
      overlay.showPendingTransaction({...mockTransaction, amount: '1.0'});
      
      // Wait a bit for any animations to settle
      await new Promise(resolve => setTimeout(resolve, 50));
      
      const overlayElements = document.querySelectorAll('.algonius-transaction-overlay');
      
      // Should only have one overlay (replaces previous)
      expect(overlayElements.length).toBe(1);
      expect(overlayElements[0].innerHTML).toContain('<strong>Amount:</strong> 1.0 ETH');
    });

    it('should handle hideOverlay when no overlay exists', () => {
      // Should not throw error
      expect(() => overlay.hideOverlay()).not.toThrow();
    });

    it('should handle multiple calls to hideOverlay', async () => {
      overlay.showPendingTransaction(mockTransaction);
      overlay.hideOverlay();
      
      // Should not throw error on second call
      expect(() => overlay.hideOverlay()).not.toThrow();
      
      // Wait for animation to complete
      await new Promise(resolve => setTimeout(resolve, 350));
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay');
      expect(overlayElement).toBeNull();
    });

    it('should clear internal overlay reference after hiding', async () => {
      overlay.showPendingTransaction(mockTransaction);
      expect(overlay['overlay']).toBeInstanceOf(HTMLElement);
      
      overlay.hideOverlay();
      
      // Wait for animation to complete
      await new Promise(resolve => setTimeout(resolve, 350));
      
      expect(overlay['overlay']).toBeNull();
    });

    it('should handle DOM element removal if overlay is removed externally', () => {
      overlay.showPendingTransaction(mockTransaction);
      
      // Simulate external removal
      const overlayElement = document.querySelector('.algonius-transaction-overlay');
      overlayElement?.remove();
      
      // hideOverlay should still work without error
      expect(() => overlay.hideOverlay()).not.toThrow();
    });
  });
});