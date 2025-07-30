import { test as base, chromium, type BrowserContext } from '@playwright/test';
import path from 'path';
import { fileURLToPath } from 'url';
import { existsSync, readFileSync, writeFileSync, mkdirSync } from 'fs';
import { homedir } from 'os';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Function to update manifest file with correct extension ID
function updateManifestFile(extensionId: string) {
  // Check multiple possible locations for the native messaging manifest
  const possiblePaths = [
    path.join(homedir(), '.config', 'google-chrome', 'NativeMessagingHosts', 'ai.algonius.wallet.json'),
    path.join(homedir(), 'Library', 'Application Support', 'Google', 'Chrome', 'NativeMessagingHosts', 'ai.algonius.wallet.json'),
    path.join(homedir(), '.algonius-wallet', 'ai.algonius.wallet.json'), // CI location
    path.join(homedir(), '.algonius-wallet', 'com.algonius.wallet.json'), // Legacy location
  ];
  
  let manifestPath = possiblePaths.find(p => existsSync(p));
  
  // If no manifest exists, try to create one in the appropriate location
  if (!manifestPath) {
    // Use CI location for CI environment, macOS location otherwise
    if (process.env.CI) {
      manifestPath = possiblePaths[0];
      // Ensure directory exists
      const dir = path.dirname(manifestPath);
      if (!existsSync(dir)) {
        mkdirSync(dir, { recursive: true });
      }
    } else {
      manifestPath = possiblePaths[1];
    }
  }
  
  try {
    // Update all existing manifest files with the correct extension ID
    const manifestsToUpdate = possiblePaths.filter(p => existsSync(p));
    
    if (manifestsToUpdate.length === 0) {
      // Create a new manifest if none exist
      manifestPath = process.env.CI ? possiblePaths[0] : possiblePaths[1];
      const dir = path.dirname(manifestPath);
      if (!existsSync(dir)) {
        mkdirSync(dir, { recursive: true });
      }
      manifestsToUpdate.push(manifestPath);
    }
    
    for (const manifestPath of manifestsToUpdate) {
      let manifestContent;
      
      if (existsSync(manifestPath)) {
        manifestContent = JSON.parse(readFileSync(manifestPath, 'utf-8'));
      } else {
        // Create a new manifest
        manifestContent = {
          "name": "ai.algonius.wallet",
          "description": "Algonius Wallet Native Host",
          "path": path.join(homedir(), ".algonius-wallet", "bin", "algonius-wallet-host"),
          "type": "stdio",
          "allowed_origins": []
        };
      }
      
      // Update the allowed_origins with the correct extension ID
      manifestContent.allowed_origins = [`chrome-extension://${extensionId}/`];
      
      // Write the updated manifest back to the file
      writeFileSync(manifestPath, JSON.stringify(manifestContent, null, 2), 'utf-8');
      console.log(`Manifest updated with extension ID: ${extensionId} at ${manifestPath}`);
    }
  } catch (error) {
    console.error('Failed to update manifest file:', error);
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