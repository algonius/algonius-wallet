import { test as base, chromium, type BrowserContext } from '@playwright/test';
import path from 'path';
import { fileURLToPath } from 'url';
import { spawn, execSync } from 'child_process';
import { existsSync } from 'fs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Function to install and start native host
async function setupNativeHost() {
  try {
    // Build native host (adjust path to project root)
    const projectRoot = path.join(__dirname, '../..');
    execSync(`cd "${projectRoot}" && cd native && make build`, { stdio: 'inherit' });
    
    // Install native host to user directory
    execSync(`cd "${projectRoot}" && cd native && make install`, { stdio: 'inherit' });
    
    console.log('Native host installed successfully');
    return true;
  } catch (error) {
    console.error('Failed to install native host:', error);
    return false;
  }
}

// Function to start native host process
function startNativeHostProcess() {
  return new Promise((resolve, reject) => {
    try {
      // Check if native host binary exists
      const nativeHostPath = path.join(process.env.HOME || '', '.algonius-wallet', 'bin', 'algonius-wallet-host');
      if (!existsSync(nativeHostPath)) {
        reject(new Error('Native host binary not found'));
        return;
      }

      // Start native host process
      const nativeHost = spawn(nativeHostPath, [], {
        stdio: ['pipe', 'pipe', 'pipe']
      });

      nativeHost.stdout?.on('data', (data) => {
        console.log(`Native Host stdout: ${data}`);
      });

      nativeHost.stderr?.on('data', (data) => {
        console.log(`Native Host stderr: ${data}`);
      });

      nativeHost.on('error', (error) => {
        console.error('Failed to start native host:', error);
        reject(error);
      });

      // Give it a moment to start
      setTimeout(() => {
        resolve(nativeHost);
      }, 2000);
    } catch (error) {
      reject(error);
    }
  });
}

export const test = base.extend<{
  context: BrowserContext;
  extensionId: string;
  nativeHost: any;
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
    await use(extensionId);
  },
  nativeHost: async ({}, use) => {
    // Try to set up and start native host
    let nativeHostProcess = null;
    
    try {
      // Install native host if needed
      const installed = await setupNativeHost();
      if (installed) {
        // Start native host process
        nativeHostProcess = await startNativeHostProcess();
        console.log('Native host process started');
      } else {
        console.log('Skipping native host setup');
      }
    } catch (error) {
      console.log('Native host setup failed:', error);
    }
    
    await use(nativeHostProcess);
    
    // Cleanup
    if (nativeHostProcess) {
      try {
        nativeHostProcess.kill();
        console.log('Native host process stopped');
      } catch (error) {
        console.error('Failed to stop native host process:', error);
      }
    }
  },
});

export const expect = test.expect;