import React from 'react';

export interface ButtonProps {
  children: React.ReactNode;
  onClick?: () => void;
  disabled?: boolean;
  variant?: 'primary' | 'secondary' | 'danger' | 'success';
  size?: 'small' | 'medium' | 'large';
  fullWidth?: boolean;
  loading?: boolean;
  type?: 'button' | 'submit' | 'reset';
  className?: string;
}

export const Button: React.FC<ButtonProps> = ({
  children,
  onClick,
  disabled = false,
  variant = 'primary',
  size = 'medium',
  fullWidth = false,
  loading = false,
  type = 'button',
  className = '',
  ...props
}) => {
  const baseClasses = [
    'font-medium',
    'rounded',
    'transition-colors',
    'focus:outline-none',
    'focus:ring-2',
    'focus:ring-offset-2',
    'disabled:opacity-50',
    'disabled:cursor-not-allowed',
    'flex',
    'items-center',
    'justify-center',
    'gap-2'
  ];

  const variantClasses = {
    primary: [
      'bg-blue-500',
      'hover:bg-blue-600',
      'text-white',
      'focus:ring-blue-500',
      'disabled:hover:bg-blue-500'
    ],
    secondary: [
      'bg-gray-200',
      'hover:bg-gray-300',
      'text-gray-700',
      'focus:ring-gray-500',
      'disabled:hover:bg-gray-200'
    ],
    danger: [
      'bg-red-500',
      'hover:bg-red-600',
      'text-white',
      'focus:ring-red-500',
      'disabled:hover:bg-red-500'
    ],
    success: [
      'bg-green-500',
      'hover:bg-green-600',
      'text-white',
      'focus:ring-green-500',
      'disabled:hover:bg-green-500'
    ]
  };

  const sizeClasses = {
    small: ['px-3', 'py-1', 'text-sm'],
    medium: ['px-4', 'py-2', 'text-sm'],
    large: ['px-6', 'py-3', 'text-base']
  };

  const widthClasses = fullWidth ? ['w-full'] : [];

  const allClasses = [
    ...baseClasses,
    ...variantClasses[variant],
    ...sizeClasses[size],
    ...widthClasses,
    className
  ].join(' ');

  return (
    <button
      type={type}
      className={allClasses}
      onClick={onClick}
      disabled={disabled || loading}
      {...props}
    >
      {loading && (
        <svg
          className="animate-spin h-4 w-4"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      )}
      {children}
    </button>
  );
};