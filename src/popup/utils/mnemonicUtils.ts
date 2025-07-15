// Mnemonic utilities for wallet operations

/**
 * Generates a secure random mnemonic phrase
 * NOTE: This is a placeholder implementation. In production, this should use
 * a proper BIP39 library with cryptographically secure random generation
 */
export function generateMnemonic(): string {
  // This is a simplified implementation for demonstration
  // In production, use a proper BIP39 library like bip39 or ethers.js
  const words = [
    'abandon', 'ability', 'able', 'about', 'above', 'absent', 'absorb', 'abstract',
    'absurd', 'abuse', 'access', 'accident', 'account', 'accuse', 'achieve', 'acid',
    'acoustic', 'acquire', 'across', 'act', 'action', 'actor', 'actress', 'actual',
    'adapt', 'add', 'addict', 'address', 'adjust', 'admit', 'adult', 'advance'
  ];
  
  const mnemonic: string[] = [];
  for (let i = 0; i < 12; i++) {
    const randomIndex = Math.floor(Math.random() * words.length);
    mnemonic.push(words[randomIndex]);
  }
  
  return mnemonic.join(' ');
}

/**
 * Copies text to clipboard
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch (err) {
    // Fallback for older browsers
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    
    try {
      document.execCommand('copy');
      document.body.removeChild(textArea);
      return true;
    } catch (err) {
      document.body.removeChild(textArea);
      return false;
    }
  }
}

/**
 * Formats mnemonic for display in a grid layout
 */
export function formatMnemonicGrid(mnemonic: string): Array<{ index: number; word: string }> {
  const words = mnemonic.trim().split(/\s+/);
  return words.map((word, index) => ({
    index: index + 1,
    word: word
  }));
}

/**
 * Shuffles mnemonic words for backup verification
 */
export function shuffleMnemonicWords(mnemonic: string): Array<{ index: number; word: string }> {
  const words = formatMnemonicGrid(mnemonic);
  const shuffled = [...words];
  
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  
  return shuffled;
}

/**
 * Verifies if user has correctly arranged mnemonic words
 */
export function verifyMnemonicOrder(
  originalMnemonic: string,
  userArrangedWords: Array<{ index: number; word: string }>
): boolean {
  const original = formatMnemonicGrid(originalMnemonic);
  
  if (original.length !== userArrangedWords.length) {
    return false;
  }
  
  for (let i = 0; i < original.length; i++) {
    if (original[i].word !== userArrangedWords[i].word) {
      return false;
    }
  }
  
  return true;
}

/**
 * Masks mnemonic for display (showing only first and last few characters)
 */
export function maskMnemonic(mnemonic: string): string {
  const words = mnemonic.trim().split(/\s+/);
  return words.map(word => {
    if (word.length <= 4) {
      return word.charAt(0) + '*'.repeat(word.length - 1);
    }
    return word.charAt(0) + '*'.repeat(word.length - 2) + word.charAt(word.length - 1);
  }).join(' ');
}

/**
 * Validates mnemonic word count
 */
export function isValidMnemonicLength(mnemonic: string): boolean {
  const words = mnemonic.trim().split(/\s+/);
  return words.length === 12 || words.length === 24;
}

/**
 * Normalizes mnemonic input (removes extra spaces, converts to lowercase)
 */
export function normalizeMnemonic(mnemonic: string): string {
  return mnemonic
    .trim()
    .toLowerCase()
    .replace(/\s+/g, ' ');
}