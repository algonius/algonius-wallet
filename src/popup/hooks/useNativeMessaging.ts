// Hook for native messaging communication

import { useCallback } from 'react';
import { WalletCreationParams, WalletImportParams, WalletOperationResponse, WalletCreationResult, WalletImportResult } from '../types/wallet';

/**
 * Hook for communicating with the native messaging system
 */
export function useNativeMessaging() {
  /**
   * Sends an RPC request to the native host via background script
   */
  const sendRpcRequest = useCallback(async (method: string, params: unknown): Promise<unknown> => {
    return new Promise((resolve, reject) => {
      chrome.runtime.sendMessage({
        action: 'native_rpc',
        method,
        params
      }, (response) => {
        if (chrome.runtime.lastError) {
          reject(new Error(chrome.runtime.lastError.message));
          return;
        }
        
        if (response?.error) {
          reject(new Error(response.error.message || 'Unknown error'));
          return;
        }
        
        resolve(response?.result || response);
      });
    });
  }, []);

  /**
   * Creates a new wallet
   */
  const createWallet = useCallback(async (params: WalletCreationParams): Promise<WalletOperationResponse> => {
    try {
      const response = await sendRpcRequest('create_wallet', params);
      return {
        success: true,
        result: response as WalletCreationResult
      };
    } catch {
      return {
        success: false,
        error: {
          code: -1,
          message: 'Unknown error'
        }
      };
    }
  }, [sendRpcRequest]);

  /**
   * Imports an existing wallet
   */
  const importWallet = useCallback(async (params: WalletImportParams): Promise<WalletOperationResponse> => {
    try {
      const response = await sendRpcRequest('import_wallet', params);
      return {
        success: true,
        result: response as WalletImportResult
      };
    } catch {
      return {
        success: false,
        error: {
          code: -1,
          message: 'Unknown error'
        }
      };
    }
  }, [sendRpcRequest]);

  /**
   * Checks if native messaging is available
   */
  const isNativeMessagingAvailable = useCallback((): boolean => {
    return typeof chrome !== 'undefined' && 
           chrome.runtime && 
           chrome.runtime.sendMessage !== undefined;
  }, []);

  /**
   * Unlocks an existing wallet with password
   */
  const unlockWallet = useCallback(async (password: string): Promise<WalletOperationResponse> => {
    try {
      const response = await sendRpcRequest('unlock_wallet', { password });
      return {
        success: true,
        result: response as WalletImportResult
      };
    } catch {
      return {
        success: false,
        error: {
          code: -1,
          message: 'Unknown error'
        }
      };
    }
  }, [sendRpcRequest]);

  /**
   * Locks the wallet (clears sensitive data from memory)
   */
  const lockWallet = useCallback(async (): Promise<{ success: boolean }> => {
    try {
      await sendRpcRequest('lock_wallet', {});
      return { success: true };
    } catch {
      return { success: false };
    }
  }, [sendRpcRequest]);

  /**
   * Checks wallet status (exists, unlocked, etc.)
   */
  const getWalletStatus = useCallback(async (): Promise<{
    hasWallet: boolean;
    isUnlocked: boolean;
    address?: string;
  }> => {
    try {
      const response = await sendRpcRequest('wallet_status', {});
      return response as {
        hasWallet: boolean;
        isUnlocked: boolean;
        address?: string;
      };
    } catch {
      return {
        hasWallet: false,
        isUnlocked: false
      };
    }
  }, [sendRpcRequest]);

  return {
    createWallet,
    importWallet,
    unlockWallet,
    lockWallet,
    getWalletStatus,
    isNativeMessagingAvailable,
    sendRpcRequest
  };
}