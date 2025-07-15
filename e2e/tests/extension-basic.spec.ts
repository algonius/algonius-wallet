import { test, expect } from '../fixtures/extension';
import { EXTENSION_PAGES } from '../utils/test-data';

test.describe('Basic Extension Tests', () => {
  test('Extension loads successfully', async ({ context, extensionId, page }) => {
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
      
      // Just verify the page loaded successfully (any title is fine for now)
      expect(title).toBeTruthy();
    });

    await test.step('Verify page content loads', async () => {
      // Wait for any content to appear
      await page.waitForSelector('body', { timeout: 10000 });
      
      // Check if body has content
      const bodyContent = await page.locator('body').textContent();
      expect(bodyContent).toBeTruthy();
      
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