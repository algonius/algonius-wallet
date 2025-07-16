import { test, expect } from '../fixtures/extension';
import { EXTENSION_PAGES } from '../utils/test-data';

test.describe('Wallet UI Flow Tests', () => {
  test('Navigation flow through wallet setup', async ({ extensionId, page }) => {
    await test.step('Open extension popup page', async () => {
      await page.goto(`chrome-extension://${extensionId}${EXTENSION_PAGES.popup}`);
      await page.waitForLoadState('networkidle');
      await expect(page).toHaveTitle(/Algonius Wallet/);
    });

    await test.step('Verify initial status view', async () => {
      // Should show MCP Host Status section
      await expect(page.locator('text=MCP Host Status')).toBeVisible();
      
      // Should show Wallet section with Setup Wallet button
      await expect(page.locator('text=No wallet configured')).toBeVisible();
      await expect(page.locator('button', { hasText: /Setup Wallet/i })).toBeVisible();
    });

    await test.step('Navigate to wallet setup', async () => {
      const setupButton = page.locator('button', { hasText: /Setup Wallet/i });
      await setupButton.click();
      
      // Should show wallet setup options
      await expect(page.locator('text=Setup Your Wallet')).toBeVisible();
      await expect(page.locator('button', { hasText: /Create New Wallet/i })).toBeVisible();
      await expect(page.locator('button', { hasText: /Import Existing Wallet/i })).toBeVisible();
    });

    await test.step('Navigate to create wallet', async () => {
      const createButton = page.locator('button', { hasText: /Create New Wallet/i });
      await createButton.click();
      
      // Should show create wallet flow
      await expect(page.locator('text=Create New Wallet')).toBeVisible();
      await expect(page.locator('button', { hasText: /Generate New Wallet/i })).toBeVisible();
    });

    await test.step('Go back to setup', async () => {
      // Click the X button to cancel - look for close button in header
      const cancelButton = page.locator('button').filter({ 
        has: page.locator('svg').filter({ hasText: /×/ })
      });
      
      if (await cancelButton.isVisible()) {
        await cancelButton.click();
      } else {
        // Reload the page to reset to initial state
        await page.reload();
        await page.waitForLoadState('networkidle');
      }
    });

    await test.step('Navigate to import wallet', async () => {
      // We should be back at the main status view
      const setupButton = page.locator('button', { hasText: /Setup Wallet/i });
      await setupButton.click();
      
      const importButton = page.locator('button', { hasText: /Import Existing Wallet/i });
      await importButton.click();
      
      // Should show import wallet flow - use specific selector
      await expect(page.locator('h1', { hasText: 'Import Wallet' })).toBeVisible();
      await expect(page.locator('textarea')).toBeVisible();
    });
  });

  test('Wallet setup UI elements visibility', async ({ extensionId, page }) => {
    await test.step('Open extension and navigate to create wallet', async () => {
      await page.goto(`chrome-extension://${extensionId}${EXTENSION_PAGES.popup}`);
      await page.waitForLoadState('networkidle');
      
      const setupButton = page.locator('button', { hasText: /Setup Wallet/i });
      await setupButton.click();
      
      const createButton = page.locator('button', { hasText: /Create New Wallet/i });
      await createButton.click();
    });

    await test.step('Verify create wallet setup step', async () => {
      // Should show security warnings
      await expect(page.locator('text=Before you start')).toBeVisible();
      await expect(page.locator('text=Make sure you\'re in a private, secure location')).toBeVisible();
      
      // Should show generate button
      await expect(page.locator('button', { hasText: /Generate New Wallet/i })).toBeVisible();
    });

    await test.step('Generate wallet and verify mnemonic display', async () => {
      const generateButton = page.locator('button', { hasText: /Generate New Wallet/i });
      await generateButton.click();
      
      // Should show mnemonic display
      await expect(page.locator('h2', { hasText: /Your Recovery Phrase/i })).toBeVisible();
      
      // Should show action buttons
      await expect(page.locator('button', { hasText: /Back/i })).toBeVisible();
      await expect(page.locator('button', { hasText: /I've Saved It/i })).toBeVisible();
    });
  });

  test('Wallet import UI elements visibility', async ({ extensionId, page }) => {
    await test.step('Open extension and navigate to import wallet', async () => {
      await page.goto(`chrome-extension://${extensionId}${EXTENSION_PAGES.popup}`);
      await page.waitForLoadState('networkidle');
      
      const setupButton = page.locator('button', { hasText: /Setup Wallet/i });
      await setupButton.click();
      
      const importButton = page.locator('button', { hasText: /Import Existing Wallet/i });
      await importButton.click();
    });

    await test.step('Verify import wallet mnemonic step', async () => {
      // Should show import title
      await expect(page.locator('h2', { hasText: /Import Wallet/i })).toBeVisible();
      
      // Should show mnemonic input
      await expect(page.locator('textarea')).toBeVisible();
      
      // Should show security notice
      await expect(page.locator('text=Security Notice')).toBeVisible();
      await expect(page.locator('text=Never share your recovery phrase')).toBeVisible();
      
      // Should show continue button (may be disabled initially)
      await expect(page.locator('button', { hasText: /Continue/i })).toBeVisible();
    });

    await test.step('Enter some text and verify validation feedback', async () => {
      const mnemonicInput = page.locator('textarea');
      await mnemonicInput.fill('test input');
      
      // Should show word count
      await expect(page.locator('text=Word count:')).toBeVisible();
      
      // Should show validation status (use more specific selector)
      await expect(page.locator('span.font-medium.text-red-600', { hasText: 'Invalid' })).toBeVisible();
      
      // Continue button should be disabled for invalid input
      const continueButton = page.locator('button', { hasText: /Continue/i });
      await expect(continueButton).toBeDisabled();
    });
  });
});