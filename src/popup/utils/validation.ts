// Validation utilities for wallet operations

import { PasswordRequirements, MnemonicValidation } from '../types/wallet';

// BIP39 word list (simplified version for validation)
// In production, this should be imported from a complete BIP39 library
const BIP39_WORDS = [
  'abandon', 'ability', 'able', 'about', 'above', 'absent', 'absorb', 'abstract',
  'absurd', 'abuse', 'access', 'accident', 'account', 'accuse', 'achieve', 'acid',
  'acoustic', 'acquire', 'across', 'act', 'action', 'actor', 'actress', 'actual',
  'adapt', 'add', 'addict', 'address', 'adjust', 'admit', 'adult', 'advance',
  'advice', 'aerobic', 'affair', 'afford', 'afraid', 'again', 'age', 'agent',
  'agree', 'ahead', 'aim', 'air', 'airport', 'aisle', 'alarm', 'album',
  'alcohol', 'alert', 'alien', 'all', 'alley', 'allow', 'almost', 'alone',
  'alpha', 'already', 'also', 'alter', 'always', 'amateur', 'amazing', 'among',
  // ... This is just a sample. In production, use a complete BIP39 word list
  'zone', 'zoo'
];

/**
 * Validates password strength according to security requirements
 */
export function validatePassword(password: string): PasswordRequirements {
  const requirements: PasswordRequirements = {
    minLength: password.length >= 8,
    hasUppercase: /[A-Z]/.test(password),
    hasLowercase: /[a-z]/.test(password),
    hasNumber: /\d/.test(password),
    hasSpecialChar: /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password),
  };

  return requirements;
}

/**
 * Checks if password meets all requirements
 */
export function isPasswordValid(password: string): boolean {
  const requirements = validatePassword(password);
  return Object.values(requirements).every(Boolean);
}

/**
 * Validates mnemonic phrase according to BIP39 standard
 */
export function validateMnemonic(mnemonic: string): MnemonicValidation {
  const words = mnemonic.trim().toLowerCase().split(/\s+/);
  const errors: string[] = [];

  // Check word count
  if (words.length !== 12 && words.length !== 24) {
    errors.push('Mnemonic must be 12 or 24 words');
  }

  // Check if all words are valid BIP39 words
  const invalidWords = words.filter(word => !BIP39_WORDS.includes(word));
  if (invalidWords.length > 0) {
    errors.push(`Invalid words: ${invalidWords.join(', ')}`);
  }

  // Check for duplicates
  const uniqueWords = new Set(words);
  if (uniqueWords.size !== words.length) {
    errors.push('Mnemonic contains duplicate words');
  }

  // Check for empty words
  if (words.some(word => word === '')) {
    errors.push('Mnemonic contains empty words');
  }

  return {
    isValid: errors.length === 0,
    wordCount: words.length,
    errors,
  };
}

/**
 * Validates chain name
 */
export function validateChain(chain: string): boolean {
  const supportedChains = ['ethereum', 'bsc'];
  return supportedChains.includes(chain.toLowerCase());
}

/**
 * Validates derivation path format
 */
export function validateDerivationPath(path: string): boolean {
  // Basic validation for derivation path format
  const pathRegex = /^m(\/\d+'?)*$/;
  return pathRegex.test(path);
}

/**
 * Sanitizes mnemonic input by normalizing spaces and case
 */
export function sanitizeMnemonic(mnemonic: string): string {
  return mnemonic
    .trim()
    .toLowerCase()
    .replace(/\s+/g, ' ');
}

/**
 * Formats mnemonic for display (with proper spacing)
 */
export function formatMnemonicForDisplay(mnemonic: string): string {
  const words = sanitizeMnemonic(mnemonic).split(' ');
  const formatted: string[] = [];
  
  for (let i = 0; i < words.length; i += 4) {
    formatted.push(words.slice(i, i + 4).join(' '));
  }
  
  return formatted.join('\n');
}

/**
 * Generates a readable password strength message
 */
export function getPasswordStrengthMessage(password: string): string {
  const requirements = validatePassword(password);
  const failedRequirements: string[] = [];

  if (!requirements.minLength) failedRequirements.push('at least 8 characters');
  if (!requirements.hasUppercase) failedRequirements.push('an uppercase letter');
  if (!requirements.hasLowercase) failedRequirements.push('a lowercase letter');
  if (!requirements.hasNumber) failedRequirements.push('a number');
  if (!requirements.hasSpecialChar) failedRequirements.push('a special character');

  if (failedRequirements.length === 0) {
    return 'Password meets all requirements';
  }

  return `Password must contain: ${failedRequirements.join(', ')}`;
}