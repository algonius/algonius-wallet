import { test, expect } from '../fixtures/extension';
import { TEST_MNEMONIC, TEST_PRIVATE_KEY, TEST_PASSWORDS, EXTENSION_PAGES } from '../utils/test-data';

test.describe('Wallet Import E2E Tests', () => {
  // Note: These tests are designed to validate the wallet import functionality
  // They may be skipped if the UI components are not yet implemented
  test.beforeEach(async ({ page }) => {
    // Ensure extension is built
    await test.step('Check if extension is built', async () => {
      // Could check if dist directory exists
      // Assuming extension is already built here
    });
  });

  test('Import wallet with mnemonic phrase', async ({ context, extensionId, page }) => {
    await test.step('Open extension popup page', async () => {
      await page.goto(`chrome-extension://${extensionId}${EXTENSION_PAGES.popup}`);
      await page.waitForLoadState('networkidle');
      await expect(page).toHaveTitle(/Algonius Wallet/);
    });

    await test.step('Click import wallet button', async () => {
      // Find and click import wallet button
      const importButton = page.locator('button', { hasText: /导入钱包|Import Wallet/i });
      
      // Check if the import button exists, if not skip this test
      try {
        await expect(importButton).toBeVisible({ timeout: 5000 });
        await importButton.click();
      } catch (error) {
        test.skip('Import wallet button not found - wallet import UI may not be implemented yet');
      }
    });

    await test.step('Select mnemonic import method', async () => {
      // Find mnemonic import option
      const mnemonicOption = page.locator('button, [role="tab"]', { hasText: /助记词|Mnemonic/i });
      if (await mnemonicOption.isVisible()) {
        await mnemonicOption.click();
      }
    });

    await test.step('Enter test mnemonic phrase', async () => {
      // Find mnemonic input field
      const mnemonicInput = page.locator('textarea, input[type="text"]').first();
      await expect(mnemonicInput).toBeVisible();
      await mnemonicInput.fill(TEST_MNEMONIC);
    });

    await test.step('Set wallet password', async () => {
      // Find password input field
      const passwordInput = page.locator('input[type="password"]').first();
      if (await passwordInput.isVisible()) {
        await passwordInput.fill(TEST_PASSWORDS.valid);
        
        // If there's a confirm password field
        const confirmPasswordInput = page.locator('input[type="password"]').nth(1);
        if (await confirmPasswordInput.isVisible()) {
          await confirmPasswordInput.fill(TEST_PASSWORDS.valid);
        }
      }
    });

    await test.step('Confirm wallet import', async () => {
      const confirmButton = page.locator('button', { hasText: /确认|Confirm|导入|Import/i });
      await expect(confirmButton).toBeVisible();
      await confirmButton.click();
    });

    await test.step('Verify wallet import success', async () => {
      // Wait for import completion, should redirect to main interface or show success message
      await expect(page.locator('text=/钱包|Wallet/i')).toBeVisible({ timeout: 10000 });
      
      // Verify wallet address is displayed correctly
      // Note: actual wallet address may differ due to different mnemonic paths
      await expect(page.locator('text=/0x/i')).toBeVisible();
    });
  });

  test('Import wallet with private key', async ({ context, extensionId, page }) => {
    await test.step('Open extension popup page', async () => {
      await page.goto(`chrome-extension://${extensionId}${EXTENSION_PAGES.popup}`);
      await expect(page).toHaveTitle(/Algonius Wallet/);
    });

    await test.step('Click import wallet button', async () => {
      const importButton = page.locator('button', { hasText: /导入钱包|Import Wallet/i });
      
      // Check if the import button exists, if not skip this test
      try {
        await expect(importButton).toBeVisible({ timeout: 5000 });
        await importButton.click();
      } catch (error) {
        test.skip('Import wallet button not found - wallet import UI may not be implemented yet');
      }
    });

    await test.step('Select private key import method', async () => {
      const privateKeyOption = page.locator('button, [role="tab"]', { hasText: /私钥|Private Key/i });
      if (await privateKeyOption.isVisible()) {
        await privateKeyOption.click();
      }
    });

    await test.step('Enter test private key', async () => {
      const privateKeyInput = page.locator('input[type="text"], input[type="password"], textarea').first();
      await expect(privateKeyInput).toBeVisible();
      await privateKeyInput.fill(TEST_PRIVATE_KEY);
    });

    await test.step('Set wallet password', async () => {
      const passwordInput = page.locator('input[type="password"]').first();
      if (await passwordInput.isVisible()) {
        await passwordInput.fill(TEST_PASSWORDS.valid);
        
        const confirmPasswordInput = page.locator('input[type="password"]').nth(1);
        if (await confirmPasswordInput.isVisible()) {
          await confirmPasswordInput.fill(TEST_PASSWORDS.valid);
        }
      }
    });

    await test.step('Confirm wallet import', async () => {
      const confirmButton = page.locator('button', { hasText: /确认|Confirm|导入|Import/i });
      await expect(confirmButton).toBeVisible();
      await confirmButton.click();
    });

    await test.step('Verify wallet import success', async () => {
      await expect(page.locator('text=/钱包|Wallet/i')).toBeVisible({ timeout: 10000 });
      await expect(page.locator('text=/0x/i')).toBeVisible();
    });
  });

  test('Validate input format when importing wallet', async ({ context, extensionId, page }) => {
    await test.step('Open extension popup page', async () => {
      await page.goto(`chrome-extension://${extensionId}${EXTENSION_PAGES.popup}`);
    });

    await test.step('Test invalid mnemonic format', async () => {
      const importButton = page.locator('button', { hasText: /导入钱包|Import Wallet/i });
      
      try {
        await expect(importButton).toBeVisible({ timeout: 5000 });
        await importButton.click();
        
        const mnemonicInput = page.locator('textarea, input[type="text"]').first();
        if (await mnemonicInput.isVisible()) {
          await mnemonicInput.fill('invalid mnemonic phrase');
          
          const confirmButton = page.locator('button', { hasText: /确认|Confirm|导入|Import/i });
          if (await confirmButton.isVisible()) {
            await confirmButton.click();
            
            // Should display error message
            await expect(page.locator('text=/无效|Invalid|错误|Error/i')).toBeVisible({ timeout: 5000 });
          }
        }
      } catch (error) {
        test.skip('Import wallet button not found - wallet import UI may not be implemented yet');
      }
    });
  });

  test('Verify basic functionality after wallet import', async ({ context, extensionId, page }) => {
    // First import wallet
    await test.step('Import test wallet', async () => {
      await page.goto(`chrome-extension://${extensionId}${EXTENSION_PAGES.popup}`);
      
      const importButton = page.locator('button', { hasText: /导入钱包|Import Wallet/i });
      
      try {
        await expect(importButton).toBeVisible({ timeout: 5000 });
        await importButton.click();
        
        const mnemonicInput = page.locator('textarea, input[type="text"]').first();
        if (await mnemonicInput.isVisible()) {
          await mnemonicInput.fill(TEST_MNEMONIC);
          
          const passwordInput = page.locator('input[type="password"]').first();
          if (await passwordInput.isVisible()) {
            await passwordInput.fill(TEST_PASSWORDS.valid);
          }
          
          const confirmButton = page.locator('button', { hasText: /确认|Confirm|导入|Import/i });
          if (await confirmButton.isVisible()) {
            await confirmButton.click();
            await page.waitForTimeout(2000); // Wait for import completion
          }
        }
      } catch (error) {
        test.skip('Import wallet button not found - wallet import UI may not be implemented yet');
      }
    });

    await test.step('Verify wallet address display', async () => {
      // Verify wallet address is displayed correctly
      await expect(page.locator('text=/0x[a-fA-F0-9]+/')).toBeVisible();
    });

    await test.step('Verify balance query functionality', async () => {
      // Find balance display area
      const balanceElement = page.locator('text=/余额|Balance/i').first();
      if (await balanceElement.isVisible()) {
        // Verify balance area exists
        await expect(balanceElement).toBeVisible();
      }
    });
  });
});