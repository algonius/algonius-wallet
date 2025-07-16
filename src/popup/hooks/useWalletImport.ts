// Hook for wallet import operations

import { useState, useCallback } from 'react';
import { useNativeMessaging } from './useNativeMessaging';
import { WalletImportParams, WalletImportResult, SupportedChain } from '../types/wallet';
import { validateMnemonic, isPasswordValid, validateChain } from '../utils/validation';
import { sanitizeMnemonic } from '../utils/validation';

export interface WalletImportState {
  isLoading: boolean;
  error: string | null;
  result: WalletImportResult | null;
  step: 'mnemonic' | 'chain' | 'password' | 'importing' | 'success';
  mnemonic: string;
  selectedChain: SupportedChain;
  derivationPath: string;
}

/**
 * Hook for managing wallet import flow
 */
export function useWalletImport() {
  const { importWallet } = useNativeMessaging();
  
  const [state, setState] = useState<WalletImportState>({
    isLoading: false,
    error: null,
    result: null,
    step: 'mnemonic',
    mnemonic: '',
    selectedChain: 'ethereum',
    derivationPath: "m/44'/60'/0'/0/0"
  });

  /**
   * Updates the mnemonic phrase
   */
  const updateMnemonic = useCallback((mnemonic: string) => {
    const sanitized = sanitizeMnemonic(mnemonic);
    setState(prev => ({
      ...prev,
      mnemonic: sanitized,
      error: null
    }));
  }, []);

  /**
   * Validates the current mnemonic
   */
  const validateCurrentMnemonic = useCallback(() => {
    const validation = validateMnemonic(state.mnemonic);
    
    if (!validation.isValid) {
      setState(prev => ({
        ...prev,
        error: validation.errors.join(', ')
      }));
      return false;
    }
    
    setState(prev => ({
      ...prev,
      error: null
    }));
    return true;
  }, [state.mnemonic]);

  /**
   * Proceeds to chain selection step
   */
  const proceedToChainSelection = useCallback(() => {
    if (validateCurrentMnemonic()) {
      setState(prev => ({
        ...prev,
        step: 'chain'
      }));
    }
  }, [validateCurrentMnemonic]);

  /**
   * Updates the selected chain
   */
  const updateChain = useCallback((chain: SupportedChain) => {
    setState(prev => ({
      ...prev,
      selectedChain: chain,
      error: null
    }));
  }, []);

  /**
   * Proceeds to password step
   */
  const proceedToPassword = useCallback(() => {
    if (validateChain(state.selectedChain)) {
      setState(prev => ({
        ...prev,
        step: 'password'
      }));
    } else {
      setState(prev => ({
        ...prev,
        error: 'Please select a supported chain'
      }));
    }
  }, [state.selectedChain]);

  /**
   * Updates the derivation path
   */
  const updateDerivationPath = useCallback((path: string) => {
    setState(prev => ({
      ...prev,
      derivationPath: path
    }));
  }, []);

  /**
   * Imports the wallet with the given password
   */
  const importExistingWallet = useCallback(async (password: string) => {
    // Validate password
    if (!isPasswordValid(password)) {
      setState(prev => ({
        ...prev,
        error: 'Password does not meet security requirements'
      }));
      return;
    }

    // Final validation of all inputs
    if (!validateCurrentMnemonic()) {
      setState(prev => ({
        ...prev,
        step: 'mnemonic'
      }));
      return;
    }

    setState(prev => ({
      ...prev,
      isLoading: true,
      error: null,
      step: 'importing'
    }));

    try {
      const params: WalletImportParams = {
        mnemonic: state.mnemonic,
        password,
        chain: state.selectedChain,
        derivationPath: state.derivationPath
      };

      const response = await importWallet(params);
      
      if (response.success && response.result) {
        setState(prev => ({
          ...prev,
          isLoading: false,
          result: response.result as WalletImportResult,
          step: 'success'
        }));
      } else {
        setState(prev => ({
          ...prev,
          isLoading: false,
          error: response.error?.message || 'Failed to import wallet',
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
  }, [state.mnemonic, state.selectedChain, state.derivationPath, importWallet, validateCurrentMnemonic]);

  /**
   * Resets the wallet import state
   */
  const resetState = useCallback(() => {
    setState({
      isLoading: false,
      error: null,
      result: null,
      step: 'mnemonic',
      mnemonic: '',
      selectedChain: 'ethereum',
      derivationPath: "m/44'/60'/0'/0/0"
    });
  }, []);

  /**
   * Goes back to the previous step
   */
  const goBack = useCallback(() => {
    setState(prev => {
      switch (prev.step) {
        case 'chain':
          return { ...prev, step: 'mnemonic' };
        case 'password':
          return { ...prev, step: 'chain' };
        case 'importing':
          return { ...prev, step: 'password' };
        default:
          return prev;
      }
    });
  }, []);

  /**
   * Validates if the current step can proceed
   */
  const canProceed = useCallback((step: string): boolean => {
    switch (step) {
      case 'mnemonic':
        return state.mnemonic.length > 0 && validateMnemonic(state.mnemonic).isValid;
      case 'chain':
        return validateChain(state.selectedChain);
      case 'password':
        return true; // Password validation happens in the component
      default:
        return true;
    }
  }, [state.mnemonic, state.selectedChain]);

  return {
    state,
    updateMnemonic,
    validateCurrentMnemonic,
    proceedToChainSelection,
    updateChain,
    proceedToPassword,
    updateDerivationPath,
    importExistingWallet,
    resetState,
    goBack,
    canProceed
  };
}