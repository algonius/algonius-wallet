import React, { useState, useCallback } from 'react';
import { Button } from '../common/Button';
import { LoadingOverlay } from '../common/LoadingSpinner';
import { MnemonicDisplay } from './MnemonicDisplay';
import { PasswordInput } from './PasswordInput';
import { ChainSelector } from './ChainSelector';
import { useWalletCreation } from '../../hooks/useWalletCreation';
import { SupportedChain } from '../../types/wallet';

export interface CreateWalletProps {
  onComplete?: () => void;
  onCancel?: () => void;
  className?: string;
}

export const CreateWallet: React.FC<CreateWalletProps> = ({
  onComplete,
  onCancel,
  className = '',
}) => {
  const {
    state,
    generateNewMnemonic,
    confirmBackup,
    createNewWallet,
    resetState,
    goBack,
    proceedToBackup,
    canProceed,
    validatePassword
  } = useWalletCreation();

  const [selectedChain, setSelectedChain] = useState<SupportedChain>('ethereum');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isPasswordValid, setIsPasswordValid] = useState(false);

  const handleGenerateMnemonic = useCallback(() => {
    generateNewMnemonic();
  }, [generateNewMnemonic]);

  const handleProceedToBackup = useCallback(() => {
    proceedToBackup();
  }, [proceedToBackup]);

  const handleBackupConfirmed = useCallback(() => {
    confirmBackup();
  }, [confirmBackup]);

  const handleCreateWallet = useCallback(async () => {
    if (!isPasswordValid) {
      return;
    }

    await createNewWallet({
      chain: selectedChain,
      password
    });

    // Call onComplete if wallet creation was successful
    if (onComplete && state.step === 'success') {
      onComplete();
    }
  }, [createNewWallet, selectedChain, password, isPasswordValid, onComplete, state.step]);

  const handleCancel = useCallback(() => {
    resetState();
    if (onCancel) {
      onCancel();
    }
  }, [resetState, onCancel]);

  const handlePasswordValidationChange = useCallback((isValid: boolean) => {
    setIsPasswordValid(isValid);
  }, []);

  const renderStepContent = () => {
    switch (state.step) {
      case 'setup':
        return (
          <div className="space-y-6">
            <div className="text-center space-y-4">
              <div className="mx-auto w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center">
                <svg className="w-8 h-8 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                </svg>
              </div>
              <div>
                <h2 className="text-xl font-semibold text-gray-900">Create New Wallet</h2>
                <p className="text-sm text-gray-600 mt-2">
                  Generate a new wallet with a secure recovery phrase
                </p>
              </div>
            </div>

            <div className="space-y-4">
              <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                <div className="flex items-start space-x-3">
                  <svg className="w-5 h-5 text-yellow-600 mt-0.5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L3.314 16.5c-.77.833.192 2.5 1.732 2.5z" />
                  </svg>
                  <div>
                    <h3 className="text-sm font-medium text-yellow-800">Before you start</h3>
                    <ul className="mt-2 text-sm text-yellow-700 space-y-1">
                      <li>• Make sure you're in a private, secure location</li>
                      <li>• Have a pen and paper ready to write down your recovery phrase</li>
                      <li>• Never share your recovery phrase with anyone</li>
                    </ul>
                  </div>
                </div>
              </div>

              <Button
                variant="primary"
                fullWidth
                onClick={handleGenerateMnemonic}
              >
                Generate New Wallet
              </Button>
            </div>
          </div>
        );

      case 'mnemonic':
        return (
          <div className="space-y-6">
            <div className="text-center">
              <h2 className="text-xl font-semibold text-gray-900">Your Recovery Phrase</h2>
              <p className="text-sm text-gray-600 mt-2">
                This is your wallet's recovery phrase. Save it securely.
              </p>
            </div>

            <MnemonicDisplay
              mnemonic={state.mnemonic}
            />

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
                onClick={handleProceedToBackup}
                disabled={!canProceed('mnemonic')}
                className="flex-1"
              >
                I've Saved It
              </Button>
            </div>
          </div>
        );

      case 'backup':
        return (
          <div className="space-y-6">
            <div className="text-center">
              <h2 className="text-xl font-semibold text-gray-900">Confirm Backup</h2>
              <p className="text-sm text-gray-600 mt-2">
                Please confirm that you have safely stored your recovery phrase
              </p>
            </div>

            <MnemonicDisplay
              mnemonic={state.mnemonic}
              onBackupConfirmed={handleBackupConfirmed}
              showBackupConfirmation={true}
            />

            <div className="flex space-x-3">
              <Button
                variant="secondary"
                onClick={goBack}
                className="flex-1"
              >
                Back
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

            <ChainSelector
              selectedChain={selectedChain}
              onChainChange={setSelectedChain}
            />

            <PasswordInput
              value={password}
              onChange={setPassword}
              confirmPassword={confirmPassword}
              onConfirmPasswordChange={setConfirmPassword}
              onValidationChange={handlePasswordValidationChange}
              showStrengthMeter={true}
              showConfirmField={true}
            />

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
                onClick={handleCreateWallet}
                disabled={!isPasswordValid}
                className="flex-1"
              >
                Create Wallet
              </Button>
            </div>
          </div>
        );

      case 'creating':
        return (
          <LoadingOverlay message="Creating your wallet..." />
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
                <h2 className="text-xl font-semibold text-gray-900">Wallet Created!</h2>
                <p className="text-sm text-gray-600 mt-2">
                  Your wallet has been successfully created and secured.
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
                    <span className="text-green-700 font-medium">Chain:</span>
                    <span className="text-green-900 ml-2 capitalize">{selectedChain}</span>
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
          <h1 className="text-2xl font-bold text-gray-900">Create Wallet</h1>
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
          {['setup', 'mnemonic', 'backup', 'password', 'success'].map((step, index) => (
            <div
              key={step}
              className={`flex-1 h-2 rounded-full ${
                ['setup', 'mnemonic', 'backup', 'password', 'creating', 'success'].indexOf(state.step) >= index
                  ? 'bg-blue-500'
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