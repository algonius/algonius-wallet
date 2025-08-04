/**
 * Shared types for the Algonius Wallet extension
 * These types mirror the Go structs from the native host
 */

// PendingTransaction interface based on Go struct from native/pkg/wallet/pending_transaction.go
export interface PendingTransaction {
  hash: string;
  chain: string;
  from: string;
  to: string;
  amount: string;
  token: string;
  type: string;
  status: string;
  confirmations: number;
  required_confirmations: number;
  block_number?: number;
  nonce?: number;
  gas_fee: string;
  priority: string;
  estimated_confirmation_time: string;
  submitted_at: string;
  last_checked: string;
  rejected_at?: string;
  rejection_reason?: string;
  rejection_details?: string;
  rejection_audit_log_id?: string;
}

// Message types for communication between content script and native host
export interface NativeMessage {
  type: string;
  data?: any;
  transaction?: PendingTransaction;
}

// Transaction overlay event types
export type OverlayEventType = 
  | 'ALGONIUS_PENDING_TRANSACTION'
  | 'ALGONIUS_TRANSACTION_COMPLETED'
  | 'ALGONIUS_TRANSACTION_UPDATE';

// Transaction status types
export type TransactionStatus = 
  | 'pending'
  | 'processing'
  | 'confirmed'
  | 'rejected'
  | 'failed';

// Chain types supported by the wallet
export type SupportedChain = 
  | 'ethereum'
  | 'bsc'
  | 'solana'
  | 'bitcoin'
  | 'sui';

// Transaction priority levels
export type TransactionPriority = 
  | 'low'
  | 'medium'
  | 'high';