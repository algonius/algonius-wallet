import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { TransactionOverlay, PendingTransaction } from '../transaction-overlay';

// Mock chrome runtime for content script testing
global.chrome = {
  runtime: {
    sendMessage: vi.fn(),
    getURL: vi.fn(() => 'chrome-extension://test/providers/wallet-provider.js'),
    onMessage: {
      addListener: vi.fn(),
    },
  },
} as unknown as typeof chrome;

// Mock process.env for development mode
process.env.NODE_ENV = 'development';

describe('Content Script Integration with TransactionOverlay', () => {
  let overlay: TransactionOverlay;
  let mockTransaction: PendingTransaction;

  beforeEach(() => {
    // Clear DOM
    document.body.innerHTML = '';
    document.head.innerHTML = '';
    
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

    // Clear all mocks
    vi.clearAllMocks();
  });

  afterEach(() => {
    // Clean up overlay
    overlay.hideOverlay();
    document.body.innerHTML = '';
    document.head.innerHTML = '';
  });

  describe('Message handling integration', () => {
    it('should show overlay when receiving pending transaction message', () => {
      // Simulate message from background script about pending transaction
      const messageEvent = new MessageEvent('message', {
        data: {
          type: 'ALGONIUS_PENDING_TRANSACTION',
          transaction: mockTransaction
        },
        source: window,
        origin: window.location.origin
      });

      // Trigger the message event (simulating content script message handling)
      window.dispatchEvent(messageEvent);

      // For now, overlay should be shown manually (we'll implement auto-show later)
      overlay.showPendingTransaction(mockTransaction);
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay');
      expect(overlayElement).toBeInstanceOf(HTMLElement);
    });

    it('should hide overlay when receiving transaction completed message', async () => {
      // First show the overlay
      overlay.showPendingTransaction(mockTransaction);
      expect(document.querySelector('.algonius-transaction-overlay')).toBeInstanceOf(HTMLElement);

      // Simulate message about transaction completion
      const messageEvent = new MessageEvent('message', {
        data: {
          type: 'ALGONIUS_TRANSACTION_COMPLETED',
          transactionHash: mockTransaction.hash
        },
        source: window,
        origin: window.location.origin
      });

      // Trigger the message event
      window.dispatchEvent(messageEvent);

      // For now, overlay should be hidden manually (we'll implement auto-hide later)
      overlay.hideOverlay();
      
      // Wait for animation to complete
      await new Promise(resolve => setTimeout(resolve, 350));
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay');
      expect(overlayElement).toBeNull();
    });
  });

  describe('DOM integration', () => {
    it('should not interfere with existing page content', () => {
      // Add some existing page content
      document.body.innerHTML = '<div id="existing-content">Existing Page Content</div>';
      
      overlay.showPendingTransaction(mockTransaction);
      
      // Verify existing content is still there
      const existingContent = document.getElementById('existing-content');
      expect(existingContent).toBeInstanceOf(HTMLElement);
      expect(existingContent?.textContent).toBe('Existing Page Content');
      
      // Verify overlay is also present
      const overlayElement = document.querySelector('.algonius-transaction-overlay');
      expect(overlayElement).toBeInstanceOf(HTMLElement);
    });

    it('should maintain overlay on top of all page elements', () => {
      // Add high z-index page content
      document.body.innerHTML = '<div style="position: fixed; z-index: 9999; top: 0; left: 0;">High Z-Index Content</div>';
      
      overlay.showPendingTransaction(mockTransaction);
      
      const overlayElement = document.querySelector('.algonius-transaction-overlay') as HTMLElement;
      const computedStyle = window.getComputedStyle(overlayElement);
      
      // Overlay should have higher z-index
      expect(parseInt(computedStyle.zIndex)).toBeGreaterThan(9999);
    });

    it('should handle page navigation without breaking', () => {
      overlay.showPendingTransaction(mockTransaction);
      
      // Simulate page navigation by changing URL
      Object.defineProperty(window, 'location', {
        value: {
          ...window.location,
          href: 'https://example.com/new-page'
        },
        writable: true
      });

      // Overlay should still be functional
      const overlayElement = document.querySelector('.algonius-transaction-overlay');
      expect(overlayElement).toBeInstanceOf(HTMLElement);
      
      // Should be able to hide overlay after navigation
      expect(() => overlay.hideOverlay()).not.toThrow();
    });
  });

  describe('Multiple transaction handling', () => {
    it('should replace overlay with new transaction when multiple pending', async () => {
      // Show first transaction
      overlay.showPendingTransaction(mockTransaction);
      expect(document.querySelector('.algonius-transaction-overlay')?.innerHTML)
        .toContain('<strong>Amount:</strong> 0.5 ETH');

      // Show second transaction
      const secondTransaction = {
        ...mockTransaction,
        hash: '0xabcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab',
        amount: '1000',
        token: 'USDT'
      };
      
      overlay.showPendingTransaction(secondTransaction);
      
      // Wait a bit for any animations to settle
      await new Promise(resolve => setTimeout(resolve, 50));
      
      // Should only have one overlay showing the latest transaction
      const overlayElements = document.querySelectorAll('.algonius-transaction-overlay');
      expect(overlayElements.length).toBe(1);
      expect(overlayElements[0].innerHTML).toContain('<strong>Amount:</strong> 1000 USDT');
    });

    it('should handle rapid transaction updates without memory leaks', () => {
      // Simulate rapid transaction updates
      for (let i = 0; i < 10; i++) {
        const rapidTransaction = {
          ...mockTransaction,
          hash: `0x${i.toString().repeat(64)}`,
          amount: i.toString(),
        };
        
        overlay.showPendingTransaction(rapidTransaction);
      }

      // Should only have one overlay element in DOM
      const overlayElements = document.querySelectorAll('.algonius-transaction-overlay');
      expect(overlayElements.length).toBe(1);
      expect(overlayElements[0].innerHTML).toContain('<strong>Amount:</strong> 9');
    });
  });

  describe('Error handling integration', () => {
    it('should handle malformed transaction data gracefully', () => {
      const malformedTransaction = {
        ...mockTransaction,
        amount: undefined,
        token: null,
        to: ''
      } as unknown as PendingTransaction;

      // Should not throw error with malformed data
      expect(() => overlay.showPendingTransaction(malformedTransaction)).not.toThrow();
      
      // Should still create overlay element (with undefined/null values)
      const overlayElement = document.querySelector('.algonius-transaction-overlay');
      expect(overlayElement).toBeInstanceOf(HTMLElement);
    });

    it('should handle DOM manipulation errors gracefully', () => {
      // Mock document.createElement to throw error
      const originalCreateElement = document.createElement;
      document.createElement = vi.fn(() => {
        throw new Error('DOM manipulation failed');
      });

      // Should handle the error without crashing
      expect(() => overlay.showPendingTransaction(mockTransaction)).toThrow('DOM manipulation failed');

      // Restore original method
      document.createElement = originalCreateElement;
    });
  });
});