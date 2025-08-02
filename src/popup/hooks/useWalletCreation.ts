// Hook for wallet creation operations

import { useCallback, useState } from 'react';
import { WalletCreationParams, WalletCreationResult } from '../types/wallet';
import { isPasswordValid, validatePassword } from '../utils/validation';
import { useNativeMessaging } from './useNativeMessaging';

export interface WalletCreationState {
  isLoading: boolean;
  error: string | null;
  result: WalletCreationResult | null;
  step: 'setup' | 'password' | 'creating' | 'mnemonic' | 'backup' | 'success';
  mnemonic: string;
  isBackupConfirmed: boolean;
}

/**
 * Hook for managing wallet creation flow
 */
export function useWalletCreation() {
  const { createWallet } = useNativeMessaging();
  
  const [state, setState] = useState<WalletCreationState>({
    isLoading: false,
    error: null,
    result: null,
    step: 'setup',
    mnemonic: '',
    isBackupConfirmed: false
  });

  /**
   * Moves to password step
   */
  const proceedToPassword = useCallback(() => {
    setState(prev => ({
      ...prev,
      step: 'password',
      error: null
    }));
  }, []);

  /**
   * Confirms that the user has backed up their mnemonic
   */
  const confirmBackup = useCallback(() => {
    setState(prev => ({
      ...prev,
      isBackupConfirmed: true,
      step: 'success'
    }));
  }, []);

  /**
   * Creates the wallet with the given parameters
   */
  const createNewWallet = useCallback(async (params: WalletCreationParams) => {
    // Validate password
    if (!isPasswordValid(params.password)) {
      setState(prev => ({
        ...prev,
        error: 'Password does not meet security requirements'
      }));
      return;
    }

    setState(prev => ({
      ...prev,
      isLoading: true,
      error: null,
      step: 'creating'
    }));

    try {
      const response = await createWallet(params);
      
      if (response.success && response.result) {
        const result = response.result as WalletCreationResult;
        setState(prev => ({
          ...prev,
          isLoading: false,
          result: result,
          mnemonic: result.mnemonic || '', // Use mnemonic from backend
          step: 'mnemonic' // Show mnemonic for backup
        }));
      } else {
        setState(prev => ({
          ...prev,
          isLoading: false,
          error: response.error?.message || 'Failed to create wallet',
          step: 'password'
        }));
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        isLoading: false,
        error: error instanceof Error ? error.message : 'Unknown error occurred',
        step: 'password'
      }));
    }
  }, [createWallet]);

  /**
   * Resets the wallet creation state
   */
  const resetState = useCallback(() => {
    setState({
      isLoading: false,
      error: null,
      result: null,
      step: 'setup',
      mnemonic: '',
      isBackupConfirmed: false
    });
  }, []);

  /**
   * Goes back to the previous step
   */
  const goBack = useCallback(() => {
    setState(prev => {
      switch (prev.step) {
        case 'password':
          return { ...prev, step: 'setup' };
        case 'creating':
          return { ...prev, step: 'password' };
        case 'mnemonic':
          return { ...prev, step: 'password', mnemonic: '', result: null };
        case 'backup':
          return { ...prev, step: 'mnemonic' };
        default:
          return prev;
      }
    });
  }, []);

  /**
   * Proceeds to backup confirmation step
   */
  const proceedToBackup = useCallback(() => {
    setState(prev => ({
      ...prev,
      step: 'backup'
    }));
  }, []);

  /**
   * Validates if the current step can proceed
   */
  const canProceed = useCallback((step: string, data?: { password?: string }): boolean => {
    switch (step) {
      case 'setup':
        return true;
      case 'password':
        return Boolean(data?.password && isPasswordValid(data.password));
      case 'mnemonic':
        return state.mnemonic.length > 0;
      case 'backup':
        return state.isBackupConfirmed;
      default:
        return true;
    }
  }, [state.mnemonic, state.isBackupConfirmed]);

  const generateNewMnemonic = ()=>{}

  return {
    state,
    proceedToPassword,
    confirmBackup,
    createNewWallet,
    resetState,
    goBack,
    proceedToBackup,
    canProceed,
    validatePassword,
    generateNewMnemonic
  };
}