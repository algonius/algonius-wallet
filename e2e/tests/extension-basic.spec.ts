import { test, expect } from '../fixtures/extension';
import { EXTENSION_PAGES } from '../utils/test-data';

test.describe('Basic Extension Tests', () => {
  test('Extension loads successfully', async ({ context: _context, extensionId, page }) => {
    await test.step('Open extension popup page', async () => {
      await page.goto(`chrome-extension://${extensionId}${EXTENSION_PAGES.popup}`);
      
      // Wait for page to load
      await page.waitForLoadState('networkidle');
      
      // Check if page loads without errors
      await expect(page).toHaveURL(`chrome-extension://${extensionId}${EXTENSION_PAGES.popup}`);
    });

    await test.step('Verify page title', async () => {
      // Check if title contains expected content
      const title = await page.title();
      console.log('Page title:', title);
      
      // Verify it contains "Algonius Wallet"
      expect(title).toContain('Algonius Wallet');
    });

    await test.step('Verify page content loads', async () => {
      // Wait for the main app header to appear
      await page.waitForSelector('h1', { timeout: 10000 });
      
      // Check if the main title is present
      const mainTitle = await page.locator('h1').textContent();
      expect(mainTitle).toContain('Algonius Wallet');
      
      // Check if MCP Host Status section is present
      const mcpStatusSection = page.locator('text=MCP Host Status');
      await expect(mcpStatusSection).toBeVisible();
      
      console.log('Page content loaded successfully');
    });
  });

  test('Extension service worker is running', async ({ context, extensionId }) => {
    await test.step('Check service worker', async () => {
      // Get service workers
      const serviceWorkers = context.serviceWorkers();
      
      // Should have at least one service worker (the extension background script)
      expect(serviceWorkers.length).toBeGreaterThan(0);
      
      // Check if any service worker is from our extension
      const extensionWorker = serviceWorkers.find(worker => 
        worker.url().includes(extensionId)
      );
      
      expect(extensionWorker).toBeTruthy();
      console.log('Extension service worker is running:', extensionWorker?.url());
    });
  });
});