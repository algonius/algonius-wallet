// Hook for wallet creation operations

import { useState, useCallback } from 'react';
import { useNativeMessaging } from './useNativeMessaging';
import { WalletCreationParams, WalletCreationResult, SupportedChain } from '../types/wallet';
import { validatePassword, isPasswordValid } from '../utils/validation';
import { generateMnemonic } from '../utils/mnemonicUtils';

export interface WalletCreationState {
  isLoading: boolean;
  error: string | null;
  result: WalletCreationResult | null;
  step: 'setup' | 'mnemonic' | 'backup' | 'password' | 'creating' | 'success';
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
   * Generates a new mnemonic phrase
   */
  const generateNewMnemonic = useCallback(() => {
    const newMnemonic = generateMnemonic();
    setState(prev => ({
      ...prev,
      mnemonic: newMnemonic,
      step: 'mnemonic',
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
      step: 'password'
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
        setState(prev => ({
          ...prev,
          isLoading: false,
          result: response.result as WalletCreationResult,
          step: 'success'
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
        case 'mnemonic':
          return { ...prev, step: 'setup', mnemonic: '' };
        case 'backup':
          return { ...prev, step: 'mnemonic' };
        case 'password':
          return { ...prev, step: 'backup', isBackupConfirmed: false };
        case 'creating':
          return { ...prev, step: 'password' };
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
  const canProceed = useCallback((step: string, data?: any): boolean => {
    switch (step) {
      case 'mnemonic':
        return state.mnemonic.length > 0;
      case 'backup':
        return state.isBackupConfirmed;
      case 'password':
        return data?.password && isPasswordValid(data.password);
      default:
        return true;
    }
  }, [state.mnemonic, state.isBackupConfirmed]);

  return {
    state,
    generateNewMnemonic,
    confirmBackup,
    createNewWallet,
    resetState,
    goBack,
    proceedToBackup,
    canProceed,
    validatePassword
  };
}