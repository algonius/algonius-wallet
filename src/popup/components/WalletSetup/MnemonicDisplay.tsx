import React, { useState, useCallback } from 'react';
import { Button } from '../common/Button';
import { formatMnemonicGrid, copyToClipboard } from '../../utils/mnemonicUtils';

export interface MnemonicDisplayProps {
  mnemonic: string;
  onBackupConfirmed?: () => void;
  showBackupConfirmation?: boolean;
  className?: string;
}

export const MnemonicDisplay: React.FC<MnemonicDisplayProps> = ({
  mnemonic,
  onBackupConfirmed,
  showBackupConfirmation = false,
  className = '',
}) => {
  const [copied, setCopied] = useState(false);
  const [isBackupConfirmed, setIsBackupConfirmed] = useState(false);
  const [showWarning, setShowWarning] = useState(true);

  const mnemonicWords = formatMnemonicGrid(mnemonic);

  const handleCopy = useCallback(async () => {
    const success = await copyToClipboard(mnemonic);
    if (success) {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  }, [mnemonic]);

  const handleBackupConfirmation = useCallback(() => {
    setIsBackupConfirmed(true);
    if (onBackupConfirmed) {
      onBackupConfirmed();
    }
  }, [onBackupConfirmed]);

  const dismissWarning = useCallback(() => {
    setShowWarning(false);
  }, []);

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Security Warning */}
      {showWarning && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-3">
          <div className="flex items-start space-x-2">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L3.314 16.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
            </div>
            <div className="flex-1">
              <h3 className="text-sm font-medium text-red-800">Important Security Notice</h3>
              <ul className="mt-2 text-sm text-red-700 space-y-1">
                <li>• Never share your recovery phrase with anyone</li>
                <li>• Store it in a secure location offline</li>
                <li>• Anyone with this phrase can access your wallet</li>
                <li>• We cannot recover your wallet if you lose this phrase</li>
              </ul>
            </div>
            <button
              onClick={dismissWarning}
              className="flex-shrink-0 text-red-400 hover:text-red-600"
            >
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      )}

      {/* Mnemonic Title */}
      <div className="text-center space-y-2">
        <h3 className="text-lg font-semibold text-gray-900">Your Recovery Phrase</h3>
        <p className="text-sm text-gray-600">
          Write down these {mnemonicWords.length} words in order and store them safely.
        </p>
      </div>

      {/* Mnemonic Grid */}
      <div className="bg-gray-50 rounded-lg p-4 border-2 border-dashed border-gray-300">
        <div className="grid grid-cols-2 gap-2">
          {mnemonicWords.map((wordData) => (
            <div
              key={wordData.index}
              className="flex items-center space-x-2 p-2 bg-white rounded border"
            >
              <span className="text-xs text-gray-500 font-mono w-6 text-right">
                {wordData.index}.
              </span>
              <span className="font-medium text-gray-900 font-mono">
                {wordData.word}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Copy Button */}
      <div className="flex justify-center">
        <Button
          variant="secondary"
          size="small"
          onClick={handleCopy}
          disabled={copied}
        >
          {copied ? (
            <>
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              Copied!
            </>
          ) : (
            <>
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
              Copy to Clipboard
            </>
          )}
        </Button>
      </div>

      {/* Backup Confirmation */}
      {showBackupConfirmation && (
        <div className="space-y-4">
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <h4 className="text-sm font-medium text-blue-800 mb-2">
              Backup Confirmation Required
            </h4>
            <p className="text-sm text-blue-700">
              Please confirm that you have safely stored your recovery phrase before proceeding.
            </p>
          </div>

          <div className="flex items-center space-x-3">
            <input
              type="checkbox"
              id="backup-confirmed"
              className="h-4 w-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500"
              checked={isBackupConfirmed}
              onChange={(e) => setIsBackupConfirmed(e.target.checked)}
            />
            <label htmlFor="backup-confirmed" className="text-sm text-gray-700">
              I have safely backed up my recovery phrase
            </label>
          </div>

          <Button
            variant="primary"
            fullWidth
            onClick={handleBackupConfirmation}
            disabled={!isBackupConfirmed}
          >
            Continue
          </Button>
        </div>
      )}

      {/* Additional Tips */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
        <h4 className="text-sm font-medium text-blue-800 mb-2">Backup Tips</h4>
        <ul className="text-sm text-blue-700 space-y-1">
          <li>• Write it down on paper and store in a safe place</li>
          <li>• Consider using a hardware wallet for extra security</li>
          <li>• Never store it digitally or share it online</li>
          <li>• Keep multiple copies in different secure locations</li>
        </ul>
      </div>
    </div>
  );
};