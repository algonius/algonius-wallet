import React from 'react';
import { SupportedChain } from '../../types/wallet';

export interface ChainSelectorProps {
  selectedChain: SupportedChain;
  onChainChange: (chain: SupportedChain) => void;
  disabled?: boolean;
  className?: string;
}

interface ChainOption {
  id: SupportedChain;
  name: string;
  description: string;
  icon: string;
  color: string;
}

const CHAIN_OPTIONS: ChainOption[] = [
  {
    id: 'ethereum',
    name: 'Ethereum',
    description: 'Ethereum mainnet',
    icon: 'Ξ',
    color: 'bg-blue-500'
  },
  {
    id: 'bsc',
    name: 'BSC',
    description: 'Binance Smart Chain',
    icon: 'B',
    color: 'bg-yellow-500'
  },
  {
    id: 'solana',
    name: 'Solana',
    description: 'Solana mainnet',
    icon: '◎',
    color: 'bg-purple-500'
  }
];

export const ChainSelector: React.FC<ChainSelectorProps> = ({
  selectedChain,
  onChainChange,
  disabled = false,
  className = '',
}) => {
  return (
    <div className={`space-y-3 ${className}`}>
      <label className="block text-sm font-medium text-gray-700">
        Select Network
      </label>
      
      <div className="space-y-2">
        {CHAIN_OPTIONS.map((chain) => (
          <button
            key={chain.id}
            type="button"
            className={`
              w-full p-3 rounded-lg border-2 transition-all duration-200
              ${selectedChain === chain.id
                ? 'border-blue-500 bg-blue-50 shadow-sm'
                : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
              }
              ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
              focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
            `}
            onClick={() => !disabled && onChainChange(chain.id)}
            disabled={disabled}
          >
            <div className="flex items-center space-x-3">
              {/* Chain Icon */}
              <div
                className={`
                  w-8 h-8 rounded-full flex items-center justify-center text-white font-bold
                  ${chain.color}
                `}
              >
                {chain.icon}
              </div>
              
              {/* Chain Info */}
              <div className="flex-1 text-left">
                <div className="font-medium text-gray-900">{chain.name}</div>
                <div className="text-sm text-gray-500">{chain.description}</div>
              </div>
              
              {/* Selection Indicator */}
              <div className="flex items-center">
                <div
                  className={`
                    w-4 h-4 rounded-full border-2 flex items-center justify-center
                    ${selectedChain === chain.id
                      ? 'border-blue-500 bg-blue-500'
                      : 'border-gray-300'
                    }
                  `}
                >
                  {selectedChain === chain.id && (
                    <div className="w-2 h-2 bg-white rounded-full" />
                  )}
                </div>
              </div>
            </div>
          </button>
        ))}
      </div>
      
      {/* Additional Info */}
      <div className="text-xs text-gray-500 space-y-1">
        <p>• You can switch networks later in settings</p>
        <p>• All networks support standard ERC-20 tokens</p>
      </div>
    </div>
  );
};