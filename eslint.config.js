import tseslint from '@typescript-eslint/eslint-plugin';
import tsparser from '@typescript-eslint/parser';
import prettier from 'eslint-config-prettier';

// Flat config for ESLint 9.x+
// - Browser globals for all src files
// - Node globals for config/build files
// - Prettier for formatting

const browserGlobals = {
  window: 'readonly',
  document: 'readonly',
  chrome: 'readonly',
  console: 'readonly',
  setTimeout: 'readonly',
  clearTimeout: 'readonly',
  setInterval: 'readonly',
  MessageEvent: 'readonly',
  EventListener: 'readonly',
  HTMLElement: 'readonly',
  HTMLDivElement: 'readonly',
  HTMLButtonElement: 'readonly',
  HTMLScriptElement: 'readonly',
};

const nodeGlobals = {
  __dirname: 'readonly',
  module: 'readonly',
  require: 'readonly',
  process: 'readonly',
  console: 'readonly',
};

export default [
  {
    files: ['src/**/*.{ts,js}', 'src/**/*.{tsx,jsx}', '*.ts', '*.js'],
    languageOptions: {
      parser: tsparser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
      },
      globals: browserGlobals,
    },
    plugins: {
      '@typescript-eslint': tseslint,
    },
    rules: {
      ...tseslint.configs.recommended.rules,
      // Add any additional rules here
    },
  },
  {
    files: ['vite.config.ts', 'eslint.config.js'],
    languageOptions: {
      parser: tsparser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
      },
      globals: nodeGlobals,
    },
    plugins: {
      '@typescript-eslint': tseslint,
    },
    rules: {
      ...tseslint.configs.recommended.rules,
    },
  },
  prettier,
  {
    ignores: [
      'dist/',
      'node_modules/',
      'build/',
      'coverage/',
      '*.min.js',
      '*.bundle.js',
    ],
  },
];
