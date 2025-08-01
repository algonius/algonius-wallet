import React, { useState, useCallback } from 'react';
import { Button } from '../common/Button';
import { LoadingOverlay } from '../common/LoadingSpinner';
import { PasswordInput } from './PasswordInput';
import { useNativeMessaging } from '../../hooks/useNativeMessaging';

export interface UnlockWalletProps {
  onComplete?: () => void;
  onCancel?: () => void;
  className?: string;
}

export const UnlockWallet: React.FC<UnlockWalletProps> = ({
  onComplete,
  onCancel,
  className = '',
}) => {
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  
  const { unlockWallet } = useNativeMessaging();

  const handlePasswordChange = useCallback((value: string) => {
    setPassword(value);
    if (error) {
      setError(null);
    }
  }, [error]);

  const handleUnlock = useCallback(async () => {
    if (!password.trim()) {
      setError('Password is required');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const result = await unlockWallet(password);
      
      if (result.success) {
        onComplete?.();
      } else {
        setError(result.error?.message || 'Failed to unlock wallet');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An unexpected error occurred');
    } finally {
      setIsLoading(false);
    }
  }, [password, unlockWallet, onComplete]);

  const handleKeyPress = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !isLoading) {
      handleUnlock();
    }
  }, [handleUnlock, isLoading]);

  return (
    <div className={`space-y-6 ${className}`}>
      {isLoading && (
        <LoadingOverlay message="Unlocking your wallet..." />
      )}
      
      <div className="text-center space-y-4">
        <div className="mx-auto w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center">
          <svg className="w-8 h-8 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
        </div>
        <div>
          <h2 className="text-xl font-semibold text-gray-900">Unlock Your Wallet</h2>
          <p className="text-sm text-gray-600 mt-2">
            Enter your password to unlock your existing wallet
          </p>
        </div>
      </div>

      <div className="space-y-4">
        <PasswordInput
          label="Password"
          value={password}
          onChange={handlePasswordChange}
          placeholder="Enter your wallet password"
          showStrengthMeter={false}
          showConfirmField={false}
        />
        
        {error && (
          <div className="text-sm text-red-600 bg-red-50 border border-red-200 rounded-md p-3">
            {error}
          </div>
        )}
      </div>

      <div className="flex space-x-3">
        <Button
          variant="secondary"
          onClick={onCancel}
          disabled={isLoading}
          fullWidth
        >
          Cancel
        </Button>
        <Button
          variant="primary"
          onClick={handleUnlock}
          disabled={isLoading || !password.trim()}
          fullWidth
        >
          {isLoading ? 'Unlocking...' : 'Unlock Wallet'}
        </Button>
      </div>
    </div>
  );
};