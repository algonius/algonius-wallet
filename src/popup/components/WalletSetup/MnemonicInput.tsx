import React, { useState, useCallback, useEffect } from 'react';
import { Input } from '../common/Input';
import { validateMnemonic, sanitizeMnemonic } from '../../utils/validation';
import { MnemonicValidation } from '../../types/wallet';

export interface MnemonicInputProps {
  value: string;
  onChange: (value: string) => void;
  onValidationChange?: (isValid: boolean) => void;
  placeholder?: string;
  disabled?: boolean;
  className?: string;
}

export const MnemonicInput: React.FC<MnemonicInputProps> = ({
  value,
  onChange,
  onValidationChange,
  placeholder = 'Enter your 12 or 24-word recovery phrase...',
  disabled = false,
  className = '',
}) => {
  const [isFocused, setIsFocused] = useState(false);
  const [validation, setValidation] = useState<MnemonicValidation>({
    isValid: false,
    wordCount: 0,
    errors: []
  });

  // Validate mnemonic on value change
  useEffect(() => {
    if (value.trim()) {
      const newValidation = validateMnemonic(value);
      setValidation(newValidation);
      
      if (onValidationChange) {
        onValidationChange(newValidation.isValid);
      }
    } else {
      setValidation({
        isValid: false,
        wordCount: 0,
        errors: []
      });
      
      if (onValidationChange) {
        onValidationChange(false);
      }
    }
  }, [value, onValidationChange]);

  const handleInputChange = useCallback((e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const sanitized = sanitizeMnemonic(e.target.value);
    onChange(sanitized);
  }, [onChange]);

  const handleFocus = useCallback(() => {
    setIsFocused(true);
  }, []);

  const handleBlur = useCallback(() => {
    setIsFocused(false);
  }, []);

  const handlePaste = useCallback((e: React.ClipboardEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    e.preventDefault();
    const pastedText = e.clipboardData.getData('text');
    const sanitized = sanitizeMnemonic(pastedText);
    onChange(sanitized);
  }, [onChange]);

  const getErrorMessage = (): string | undefined => {
    if (!value.trim()) return undefined;
    if (validation.errors.length > 0) {
      return validation.errors[0]; // Show first error
    }
    return undefined;
  };

  const getWordCount = (): number => {
    return value.trim().split(/\s+/).filter(word => word.length > 0).length;
  };

  const getInputStatus = (): 'default' | 'valid' | 'invalid' => {
    if (!value.trim()) return 'default';
    return validation.isValid ? 'valid' : 'invalid';
  };

  const inputStatus = getInputStatus();
  const wordCount = getWordCount();

  return (
    <div className={`space-y-3 ${className}`}>
      {/* Input with enhanced styling */}
      <div className="relative">
        <Input
          label="Recovery Phrase"
          multiline
          rows={4}
          placeholder={placeholder}
          value={value}
          onChange={handleInputChange}
          onFocus={handleFocus}
          onBlur={handleBlur}
          onPaste={handlePaste}
          disabled={disabled}
          error={getErrorMessage()}
          className={`
            font-mono text-sm
            ${inputStatus === 'valid' ? 'border-green-300 focus:border-green-500 focus:ring-green-500' : ''}
            ${inputStatus === 'invalid' ? 'border-red-300 focus:border-red-500 focus:ring-red-500' : ''}
            ${isFocused ? 'ring-2 ring-offset-1' : ''}
          `}
        />
        
        {/* Status indicator */}
        {value.trim() && (
          <div className="absolute right-3 top-8">
            {validation.isValid ? (
              <svg className="h-5 w-5 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            ) : (
              <svg className="h-5 w-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            )}
          </div>
        )}
      </div>

      {/* Word count and validation status */}
      <div className="flex items-center justify-between text-sm">
        <div className="flex items-center space-x-2">
          <span className="text-gray-500">Word count:</span>
          <span className={`font-medium ${
            wordCount === 12 || wordCount === 24 ? 'text-green-600' : 'text-gray-600'
          }`}>
            {wordCount} {wordCount === 1 ? 'word' : 'words'}
          </span>
        </div>
        
        {value.trim() && (
          <div className="flex items-center space-x-1">
            <span className={`font-medium ${
              validation.isValid ? 'text-green-600' : 'text-red-600'
            }`}>
              {validation.isValid ? 'Valid' : 'Invalid'}
            </span>
          </div>
        )}
      </div>

      {/* Validation hints */}
      {!validation.isValid && value.trim() && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-3">
          <h4 className="text-sm font-medium text-red-800 mb-2">
            Validation Issues
          </h4>
          <ul className="text-sm text-red-700 space-y-1">
            {validation.errors.map((error, index) => (
              <li key={index}>• {error}</li>
            ))}
          </ul>
        </div>
      )}

      {/* Success message */}
      {validation.isValid && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-3">
          <div className="flex items-center space-x-2">
            <svg className="h-5 w-5 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
            <span className="text-sm font-medium text-green-800">
              Valid {wordCount}-word recovery phrase detected
            </span>
          </div>
        </div>
      )}

      {/* Help text */}
      <div className="text-xs text-gray-500 space-y-1">
        <p>• Enter your recovery phrase with spaces between words</p>
        <p>• Supports 12 or 24-word phrases</p>
        <p>• Paste from clipboard is supported</p>
        <p>• Case doesn't matter - words will be normalized</p>
      </div>
    </div>
  );
};