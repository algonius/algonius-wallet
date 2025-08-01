import { defineConfig } from 'vite';
import { resolve } from 'path';
import react from '@vitejs/plugin-react';

// If you need to copy static files (manifest, icons, html), use vite-plugin-static-copy
import { viteStaticCopy } from 'vite-plugin-static-copy';

export default defineConfig({
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    rollupOptions: {
      input: {
        background: resolve(__dirname, 'src/background/background.ts'),
        content: resolve(__dirname, 'src/content/content.ts'),
        "wallet-provider": resolve(__dirname, 'src/providers/wallet-provider.js'),
        popup: resolve(__dirname, 'src/popup/index.tsx'),
        "popup/popup": resolve(__dirname, 'src/popup/index.css'),
      },
      output: {
        entryFileNames: (chunk) => {
          if (chunk.name === 'wallet-provider') return 'providers/wallet-provider.js';
          if (chunk.name === 'content') return 'content/content.js';
          if (chunk.name === 'background') return 'background/background.js';
          if (chunk.name === 'popup') return 'popup/popup.js';
          return 'assets/[name].js';
        },
        chunkFileNames: 'assets/[name]-[hash].js',
        assetFileNames: (assetInfo) => {
          if (assetInfo.name === 'popup/popup.css') return 'popup/popup.css';
          return 'assets/[name][extname]';
        },
      },
    },
  },
  plugins: [
    react(),
    viteStaticCopy({
      targets: [
        { src: 'src/manifest.json', dest: '.' },
        { src: 'src/popup/popup.html', dest: 'popup' },
        // Add icons or other static assets as needed
      ],
    }),
  ],
});
