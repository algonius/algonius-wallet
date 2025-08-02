// Wallet operation types for the popup UI

export interface WalletCreationParams {
  chain: string;
  password: string;
}

export interface WalletImportParams {
  mnemonic: string;
  password: string;
  chain: string;
  derivationPath?: string;
}

export interface WalletResult {
  address: string;
  publicKey: string;
  createdAt?: number;
  importedAt?: number;
}

export interface WalletCreationResult extends WalletResult {
  mnemonic: string;
  createdAt: number;
}

export interface WalletImportResult extends WalletResult {
  importedAt: number;
}

export interface WalletError {
  code: number;
  message: string;
}

export interface WalletOperationResponse {
  success: boolean;
  result?: WalletCreationResult | WalletImportResult;
  error?: WalletError;
}

export type SupportedChain = 'ethereum' | 'bsc' | 'solana';

export interface ValidationError {
  field: string;
  message: string;
}

export interface PasswordRequirements {
  minLength: boolean;
  hasUppercase: boolean;
  hasLowercase: boolean;
  hasNumber: boolean;
  hasSpecialChar: boolean;
}

export interface MnemonicValidation {
  isValid: boolean;
  wordCount: number;
  errors: string[];
}