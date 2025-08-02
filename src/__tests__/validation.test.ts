import { describe, it, expect } from 'vitest';
import { validateMnemonic } from '../popup/utils/validation';

describe('Mnemonic Validation', () => {
  it('should validate a correct 12-word mnemonic with duplicate words', () => {
    const mnemonic = 'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about';
    const result = validateMnemonic(mnemonic);
    expect(result.isValid).toBe(true);
    expect(result.wordCount).toBe(12);
    // The key fix: should not contain "Mnemonic contains duplicate words" error
    expect(result.errors).not.toContain('Mnemonic contains duplicate words');
  });

  it('should reject mnemonics with wrong word count', () => {
    const mnemonic = 'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon';
    const result = validateMnemonic(mnemonic);
    expect(result.isValid).toBe(false);
    expect(result.errors).toContain('Mnemonic must be 12 or 24 words');
  });

  it('should reject mnemonics with invalid words', () => {
    const mnemonic = 'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon invalidword';
    const result = validateMnemonic(mnemonic);
    expect(result.isValid).toBe(false);
    expect(result.errors).toContain('Invalid words: invalidword');
  });

  it('should allow mnemonics with duplicate words (BIP39 compliant)', () => {
    // Test with a mnemonic that has duplicate words but is otherwise valid
    // Using words that are definitely in our limited BIP39 word list
    const mnemonic = 'about about about about about about about about about about about able';
    const result = validateMnemonic(mnemonic);
    // Should not fail due to duplicate words
    expect(result.errors).not.toContain('Mnemonic contains duplicate words');
  });
});