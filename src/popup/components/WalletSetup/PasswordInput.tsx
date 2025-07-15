import React, { useState, useCallback } from 'react';
import { Input } from '../common/Input';
import { validatePassword, getPasswordStrengthMessage, isPasswordValid } from '../../utils/validation';
import { PasswordRequirements } from '../../types/wallet';

export interface PasswordInputProps {
  label?: string;
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
  onValidationChange?: (isValid: boolean) => void;
  showStrengthMeter?: boolean;
  confirmPassword?: string;
  onConfirmPasswordChange?: (value: string) => void;
  showConfirmField?: boolean;
  required?: boolean;
  disabled?: boolean;
  className?: string;
}

export const PasswordInput: React.FC<PasswordInputProps> = ({
  label = 'Password',
  placeholder = 'Enter your password',
  value,
  onChange,
  onValidationChange,
  showStrengthMeter = true,
  confirmPassword = '',
  onConfirmPasswordChange,
  showConfirmField = true,
  required = true,
  disabled = false,
  className = '',
}) => {
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [passwordTouched, setPasswordTouched] = useState(false);
  const [confirmTouched, setConfirmTouched] = useState(false);

  const requirements = validatePassword(value);
  const passwordIsValid = isPasswordValid(value);
  const passwordsMatch = value === confirmPassword;
  const strengthMessage = getPasswordStrengthMessage(value);

  // Notify parent of validation changes
  React.useEffect(() => {
    if (onValidationChange) {
      const isValid = passwordIsValid && (!showConfirmField || passwordsMatch);
      onValidationChange(isValid);
    }
  }, [passwordIsValid, passwordsMatch, showConfirmField, onValidationChange]);

  const handlePasswordChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    onChange(e.target.value);
  }, [onChange]);

  const handleConfirmPasswordChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    if (onConfirmPasswordChange) {
      onConfirmPasswordChange(e.target.value);
    }
  }, [onConfirmPasswordChange]);

  const handlePasswordBlur = useCallback(() => {
    setPasswordTouched(true);
  }, []);

  const handleConfirmBlur = useCallback(() => {
    setConfirmTouched(true);
  }, []);

  const getPasswordError = (): string | undefined => {
    if (!passwordTouched || !value) return undefined;
    if (!passwordIsValid) return strengthMessage;
    return undefined;
  };

  const getConfirmPasswordError = (): string | undefined => {
    if (!confirmTouched || !confirmPassword) return undefined;
    if (!passwordsMatch) return 'Passwords do not match';
    return undefined;
  };

  const getStrengthColor = (): string => {
    const validCount = Object.values(requirements).filter(Boolean).length;
    if (validCount <= 2) return 'bg-red-500';
    if (validCount <= 4) return 'bg-yellow-500';
    return 'bg-green-500';
  };

  const getStrengthPercentage = (): number => {
    const validCount = Object.values(requirements).filter(Boolean).length;
    return (validCount / 5) * 100;
  };

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Password Input */}
      <div className="relative">
        <Input
          label={label}
          type={showPassword ? 'text' : 'password'}
          placeholder={placeholder}
          value={value}
          onChange={handlePasswordChange}
          onBlur={handlePasswordBlur}
          error={getPasswordError()}
          required={required}
          disabled={disabled}
          autoComplete="new-password"
        />
        
        {/* Password visibility toggle */}
        <button
          type="button"
          className="absolute right-3 top-8 text-gray-400 hover:text-gray-600"
          onClick={() => setShowPassword(!showPassword)}
        >
          {showPassword ? (
            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L8.464 8.464M14.12 14.12l1.414 1.414" />
            </svg>
          ) : (
            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
            </svg>
          )}
        </button>
      </div>

      {/* Password Strength Meter */}
      {showStrengthMeter && value && (
        <div className="space-y-2">
          <div className="flex items-center space-x-2">
            <span className="text-xs text-gray-500">Strength:</span>
            <div className="flex-1 h-2 bg-gray-200 rounded-full overflow-hidden">
              <div
                className={`h-full transition-all duration-300 ${getStrengthColor()}`}
                style={{ width: `${getStrengthPercentage()}%` }}
              />
            </div>
          </div>
          
          {/* Requirements List */}
          <div className="space-y-1">
            {[
              { key: 'minLength', label: 'At least 8 characters' },
              { key: 'hasUppercase', label: 'One uppercase letter' },
              { key: 'hasLowercase', label: 'One lowercase letter' },
              { key: 'hasNumber', label: 'One number' },
              { key: 'hasSpecialChar', label: 'One special character' },
            ].map(({ key, label }) => (
              <div key={key} className="flex items-center space-x-2">
                <span
                  className={`text-xs ${
                    requirements[key as keyof PasswordRequirements]
                      ? 'text-green-600'
                      : 'text-gray-400'
                  }`}
                >
                  {requirements[key as keyof PasswordRequirements] ? '✓' : '○'}
                </span>
                <span
                  className={`text-xs ${
                    requirements[key as keyof PasswordRequirements]
                      ? 'text-green-600'
                      : 'text-gray-500'
                  }`}
                >
                  {label}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Confirm Password Input */}
      {showConfirmField && (
        <div className="relative">
          <Input
            label="Confirm Password"
            type={showConfirmPassword ? 'text' : 'password'}
            placeholder="Confirm your password"
            value={confirmPassword}
            onChange={handleConfirmPasswordChange}
            onBlur={handleConfirmBlur}
            error={getConfirmPasswordError()}
            required={required}
            disabled={disabled}
            autoComplete="new-password"
          />
          
          {/* Confirm password visibility toggle */}
          <button
            type="button"
            className="absolute right-3 top-8 text-gray-400 hover:text-gray-600"
            onClick={() => setShowConfirmPassword(!showConfirmPassword)}
          >
            {showConfirmPassword ? (
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L8.464 8.464M14.12 14.12l1.414 1.414" />
              </svg>
            ) : (
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
            )}
          </button>
        </div>
      )}
    </div>
  );
};