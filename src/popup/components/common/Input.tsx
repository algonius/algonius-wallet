import React, { forwardRef } from 'react';

export interface InputProps {
  label?: string;
  placeholder?: string;
  value?: string;
  onChange?: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
  onBlur?: (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
  onFocus?: (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
  onPaste?: (e: React.ClipboardEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
  type?: 'text' | 'password' | 'email' | 'number';
  disabled?: boolean;
  readOnly?: boolean;
  error?: string;
  helperText?: string;
  required?: boolean;
  fullWidth?: boolean;
  multiline?: boolean;
  rows?: number;
  className?: string;
  autoComplete?: string;
  autoFocus?: boolean;
}

export const Input = forwardRef<HTMLInputElement | HTMLTextAreaElement, InputProps>(
  (
    {
      label,
      placeholder,
      value,
      onChange,
      onBlur,
      onFocus,
      onPaste,
      type = 'text',
      disabled = false,
      readOnly = false,
      error,
      helperText,
      required = false,
      fullWidth = true,
      multiline = false,
      rows = 3,
      className = '',
      autoComplete,
      autoFocus = false
    },
    ref
  ) => {
    const baseClasses = [
      'border',
      'rounded',
      'px-3',
      'py-2',
      'text-sm',
      'transition-colors',
      'focus:outline-none',
      'focus:ring-2',
      'focus:ring-offset-1',
      'disabled:opacity-50',
      'disabled:cursor-not-allowed',
      'disabled:bg-gray-100'
    ];

    const stateClasses = error
      ? [
          'border-red-300',
          'focus:border-red-500',
          'focus:ring-red-500'
        ]
      : [
          'border-gray-300',
          'focus:border-blue-500',
          'focus:ring-blue-500'
        ];

    const widthClasses = fullWidth ? ['w-full'] : [];

    const inputClasses = [
      ...baseClasses,
      ...stateClasses,
      ...widthClasses,
      className
    ].join(' ');

    return (
      <div className={fullWidth ? 'w-full' : ''}>
        {label && (
          <label className="block text-sm font-medium text-gray-700 mb-1">
            {label}
            {required && <span className="text-red-500 ml-1">*</span>}
          </label>
        )}
        {multiline ? (
          <textarea
            ref={ref as React.ForwardedRef<HTMLTextAreaElement>}
            className={inputClasses}
            placeholder={placeholder}
            value={value}
            onChange={onChange}
            onBlur={onBlur}
            onFocus={onFocus}
            onPaste={onPaste}
            disabled={disabled}
            readOnly={readOnly}
            required={required}
            autoFocus={autoFocus}
            rows={rows}
          />
        ) : (
          <input
            ref={ref as React.ForwardedRef<HTMLInputElement>}
            className={inputClasses}
            placeholder={placeholder}
            value={value}
            onChange={onChange}
            onBlur={onBlur}
            onFocus={onFocus}
            onPaste={onPaste}
            type={type}
            disabled={disabled}
            readOnly={readOnly}
            required={required}
            autoComplete={autoComplete}
            autoFocus={autoFocus}
          />
        )}
        {(error || helperText) && (
          <p className={`mt-1 text-xs ${error ? 'text-red-600' : 'text-gray-500'}`}>
            {error || helperText}
          </p>
        )}
      </div>
    );
  }
);