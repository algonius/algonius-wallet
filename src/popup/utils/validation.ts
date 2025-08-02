// Validation utilities for wallet operations

import { PasswordRequirements, MnemonicValidation } from '../types/wallet';
import * as bip39 from 'bip39';

// Complete BIP39 word list for validation
// Using the bip39 library for accurate validation

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

  // Check if all words are valid BIP39 words using the complete word list
  const invalidWords = words.filter(word => !bip39.wordlists.EN.includes(word));
  if (invalidWords.length > 0) {
    errors.push(`Invalid words: ${invalidWords.join(', ')}`);
  }

  // Check for empty words
  if (words.some(word => word === '')) {
    errors.push('Mnemonic contains empty words');
  }

  // Note: BIP39 allows duplicate words as long as the checksum is valid
  // The actual validation is done in the native backend which has a complete BIP39 implementation

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