import { test as base, chromium, type BrowserContext } from '@playwright/test';
import path from 'path';
import { fileURLToPath } from 'url';
import { existsSync, readFileSync } from 'fs';
import { homedir } from 'os';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Function to update manifest file with correct extension ID
function updateManifestFile(extensionId: string) {
  const manifestPath = path.join(homedir(), '.algonius-wallet', 'com.algonius.wallet.json');
  
  if (existsSync(manifestPath)) {
    try {
      const manifestContent = JSON.parse(readFileSync(manifestPath, 'utf-8'));
      // Update the allowed_origins with the correct extension ID
      manifestContent.allowed_origins = [`chrome-extension://${extensionId}/`];
      
      // In a real implementation, we would write this back to the file
      // But for E2E testing, we'll just log it
      console.log(`Manifest would be updated with extension ID: ${extensionId}`);
    } catch (error) {
      console.error('Failed to update manifest file:', error);
    }
  }
}

export const test = base.extend<{
  context: BrowserContext;
  extensionId: string;
}>({
  context: async ({}, use) => {
    const pathToExtension = path.join(__dirname, '../../dist');
    const context = await chromium.launchPersistentContext('', {
      headless: process.env.CI ? true : false,
      args: [
        `--disable-extensions-except=${pathToExtension}`,
        `--load-extension=${pathToExtension}`,
        '--disable-web-security',
        '--disable-background-timer-throttling',
        '--disable-backgrounding-occluded-windows',
        '--disable-renderer-backgrounding',
        '--no-sandbox',
        '--disable-setuid-sandbox'
      ],
    });
    await use(context);
    await context.close();
  },
  extensionId: async ({ context }, use) => {
    // Chrome extension pages have URLs like chrome-extension://<extension-id>/...
    let [background] = context.serviceWorkers();
    if (!background) {
      background = await context.waitForEvent('serviceworker');
    }

    const extensionId = background.url().split('/')[2];
    
    // Update manifest file with correct extension ID
    updateManifestFile(extensionId);
    
    await use(extensionId);
  },
});

export const expect = test.expect;