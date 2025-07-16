import React, { useState, useCallback } from 'react';
import { Button } from '../common/Button';
import { LoadingOverlay } from '../common/LoadingSpinner';
import { MnemonicInput } from './MnemonicInput';
import { PasswordInput } from './PasswordInput';
import { ChainSelector } from './ChainSelector';
import { Input } from '../common/Input';
import { useWalletImport } from '../../hooks/useWalletImport';
import { SupportedChain } from '../../types/wallet';

export interface ImportWalletProps {
  onComplete?: () => void;
  onCancel?: () => void;
  className?: string;
}

export const ImportWallet: React.FC<ImportWalletProps> = ({
  onComplete,
  onCancel,
  className = '',
}) => {
  const {
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
  } = useWalletImport();

  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isPasswordValid, setIsPasswordValid] = useState(false);
  const [isMnemonicValid, setIsMnemonicValid] = useState(false);

  const handleMnemonicChange = useCallback((value: string) => {
    updateMnemonic(value);
  }, [updateMnemonic]);

  const handleMnemonicValidationChange = useCallback((isValid: boolean) => {
    setIsMnemonicValid(isValid);
  }, []);

  const handleProceedToChain = useCallback(() => {
    if (validateCurrentMnemonic()) {
      proceedToChainSelection();
    }
  }, [validateCurrentMnemonic, proceedToChainSelection]);

  const handleChainChange = useCallback((chain: SupportedChain) => {
    updateChain(chain);
  }, [updateChain]);

  const handleProceedToPassword = useCallback(() => {
    proceedToPassword();
  }, [proceedToPassword]);

  const handleDerivationPathChange = useCallback((e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    updateDerivationPath(e.target.value);
  }, [updateDerivationPath]);

  const handlePasswordValidationChange = useCallback((isValid: boolean) => {
    setIsPasswordValid(isValid);
  }, []);

  const handleImportWallet = useCallback(async () => {
    if (!isPasswordValid) {
      return;
    }

    await importExistingWallet(password);

    // Call onComplete if wallet import was successful
    if (onComplete && state.step === 'success') {
      onComplete();
    }
  }, [importExistingWallet, password, isPasswordValid, onComplete, state.step]);

  const handleCancel = useCallback(() => {
    resetState();
    if (onCancel) {
      onCancel();
    }
  }, [resetState, onCancel]);

  const renderStepContent = () => {
    switch (state.step) {
      case 'mnemonic':
        return (
          <div className="space-y-6">
            <div className="text-center space-y-4">
              <div className="mx-auto w-16 h-16 bg-purple-100 rounded-full flex items-center justify-center">
                <svg className="w-8 h-8 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
              </div>
              <div>
                <h2 className="text-xl font-semibold text-gray-900">Import Wallet</h2>
                <p className="text-sm text-gray-600 mt-2">
                  Enter your recovery phrase to restore your wallet
                </p>
              </div>
            </div>

            <MnemonicInput
              value={state.mnemonic}
              onChange={handleMnemonicChange}
              onValidationChange={handleMnemonicValidationChange}
            />

            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
              <div className="flex items-start space-x-3">
                <svg className="w-5 h-5 text-yellow-600 mt-0.5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L3.314 16.5c-.77.833.192 2.5 1.732 2.5z" />
                </svg>
                <div>
                  <h3 className="text-sm font-medium text-yellow-800">Security Notice</h3>
                  <p className="mt-1 text-sm text-yellow-700">
                    Never share your recovery phrase with anyone. We will never ask for it.
                  </p>
                </div>
              </div>
            </div>

            <Button
              variant="primary"
              fullWidth
              onClick={handleProceedToChain}
              disabled={!isMnemonicValid}
            >
              Continue
            </Button>
          </div>
        );

      case 'chain':
        return (
          <div className="space-y-6">
            <div className="text-center">
              <h2 className="text-xl font-semibold text-gray-900">Select Network</h2>
              <p className="text-sm text-gray-600 mt-2">
                Choose the blockchain network for your wallet
              </p>
            </div>

            <ChainSelector
              selectedChain={state.selectedChain}
              onChainChange={handleChainChange}
            />

            {/* Advanced Options */}
            <div className="space-y-3">
              <details className="group">
                <summary className="flex items-center justify-between cursor-pointer text-sm font-medium text-gray-700 hover:text-gray-900">
                  <span>Advanced Options</span>
                  <svg className="w-4 h-4 transform transition-transform group-open:rotate-180" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                  </svg>
                </summary>
                <div className="mt-3 space-y-3">
                  <Input
                    label="Derivation Path"
                    value={state.derivationPath}
                    onChange={handleDerivationPathChange}
                    placeholder="m/44'/60'/0'/0/0"
                    helperText="Custom derivation path (leave default unless you know what you're doing)"
                  />
                </div>
              </details>
            </div>

            <div className="flex space-x-3">
              <Button
                variant="secondary"
                onClick={goBack}
                className="flex-1"
              >
                Back
              </Button>
              <Button
                variant="primary"
                onClick={handleProceedToPassword}
                disabled={!canProceed('chain')}
                className="flex-1"
              >
                Continue
              </Button>
            </div>
          </div>
        );

      case 'password':
        return (
          <div className="space-y-6">
            <div className="text-center">
              <h2 className="text-xl font-semibold text-gray-900">Set Password</h2>
              <p className="text-sm text-gray-600 mt-2">
                Choose a strong password to encrypt your wallet
              </p>
            </div>

            <PasswordInput
              value={password}
              onChange={setPassword}
              confirmPassword={confirmPassword}
              onConfirmPasswordChange={setConfirmPassword}
              onValidationChange={handlePasswordValidationChange}
              showStrengthMeter={true}
              showConfirmField={true}
            />

            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <h3 className="text-sm font-medium text-blue-800 mb-2">Import Summary</h3>
              <div className="space-y-2 text-sm">
                <div>
                  <span className="text-blue-700 font-medium">Network:</span>
                  <span className="text-blue-900 ml-2 capitalize">{state.selectedChain}</span>
                </div>
                <div>
                  <span className="text-blue-700 font-medium">Derivation Path:</span>
                  <span className="text-blue-900 ml-2 font-mono">{state.derivationPath}</span>
                </div>
                <div>
                  <span className="text-blue-700 font-medium">Words:</span>
                  <span className="text-blue-900 ml-2">{state.mnemonic.trim().split(/\s+/).length} words</span>
                </div>
              </div>
            </div>

            <div className="flex space-x-3">
              <Button
                variant="secondary"
                onClick={goBack}
                className="flex-1"
              >
                Back
              </Button>
              <Button
                variant="primary"
                onClick={handleImportWallet}
                disabled={!isPasswordValid}
                className="flex-1"
              >
                Import Wallet
              </Button>
            </div>
          </div>
        );

      case 'importing':
        return (
          <LoadingOverlay message="Importing your wallet..." />
        );

      case 'success':
        return (
          <div className="space-y-6">
            <div className="text-center space-y-4">
              <div className="mx-auto w-16 h-16 bg-green-100 rounded-full flex items-center justify-center">
                <svg className="w-8 h-8 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>
              <div>
                <h2 className="text-xl font-semibold text-gray-900">Wallet Imported!</h2>
                <p className="text-sm text-gray-600 mt-2">
                  Your wallet has been successfully imported and is ready to use.
                </p>
              </div>
            </div>

            {state.result && (
              <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                <h3 className="text-sm font-medium text-green-800 mb-2">Wallet Details</h3>
                <div className="space-y-2 text-sm">
                  <div>
                    <span className="text-green-700 font-medium">Address:</span>
                    <div className="font-mono text-green-900 break-all mt-1">
                      {state.result.address}
                    </div>
                  </div>
                  <div>
                    <span className="text-green-700 font-medium">Network:</span>
                    <span className="text-green-900 ml-2 capitalize">{state.selectedChain}</span>
                  </div>
                  <div>
                    <span className="text-green-700 font-medium">Imported:</span>
                    <span className="text-green-900 ml-2">
                      {new Date(state.result.importedAt * 1000).toLocaleString()}
                    </span>
                  </div>
                </div>
              </div>
            )}

            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <h3 className="text-sm font-medium text-blue-800 mb-2">Next Steps</h3>
              <ul className="text-sm text-blue-700 space-y-1">
                <li>• Your wallet is now ready to use</li>
                <li>• You can receive funds at your wallet address</li>
                <li>• Keep your recovery phrase safe and secure</li>
                <li>• You can switch networks in settings if needed</li>
              </ul>
            </div>

            <Button
              variant="primary"
              fullWidth
              onClick={onComplete}
            >
              Continue to Wallet
            </Button>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className={`max-w-md mx-auto p-4 ${className}`}>
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900">Import Wallet</h1>
          <button
            onClick={handleCancel}
            className="text-gray-400 hover:text-gray-600"
          >
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        {/* Progress Steps */}
        <div className="mt-4 flex items-center space-x-2">
          {['mnemonic', 'chain', 'password', 'success'].map((step, index) => (
            <div
              key={step}
              className={`flex-1 h-2 rounded-full ${
                ['mnemonic', 'chain', 'password', 'importing', 'success'].indexOf(state.step) >= index
                  ? 'bg-purple-500'
                  : 'bg-gray-200'
              }`}
            />
          ))}
        </div>
      </div>

      {/* Error Display */}
      {state.error && (
        <div className="mb-4 bg-red-50 border border-red-200 rounded-lg p-3">
          <div className="flex items-center space-x-2">
            <svg className="w-5 h-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L3.314 16.5c-.77.833.192 2.5 1.732 2.5z" />
            </svg>
            <span className="text-sm text-red-700">{state.error}</span>
          </div>
        </div>
      )}

      {/* Step Content */}
      {renderStepContent()}
    </div>
  );
};