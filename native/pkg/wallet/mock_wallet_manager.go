// SPDX-License-Identifier: Apache-2.0
package wallet

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockWalletManager is a mock implementation of IWalletManager using testify/mock
type MockWalletManager struct {
	mock.Mock
}

// CreateWallet mocks the CreateWallet method
func (m *MockWalletManager) CreateWallet(ctx context.Context, chain, password string) (address string, publicKey string, mnemonic string, err error) {
	args := m.Called(ctx, chain, password)
	return args.String(0), args.String(1), args.String(2), args.Error(3)
}

// ImportWallet mocks the ImportWallet method
func (m *MockWalletManager) ImportWallet(ctx context.Context, mnemonic, password, chainName, derivationPath string) (address string, publicKey string, importedAt int64, err error) {
	args := m.Called(ctx, mnemonic, password, chainName, derivationPath)
	return args.String(0), args.String(1), args.Get(2).(int64), args.Error(3)
}

// GetBalance mocks the GetBalance method
func (m *MockWalletManager) GetBalance(ctx context.Context, address string, token string) (string, error) {
	args := m.Called(ctx, address, token)
	return args.String(0), args.Error(1)
}

// GetStatus mocks the GetStatus method
func (m *MockWalletManager) GetStatus(ctx context.Context) (*WalletStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*WalletStatus), args.Error(1)
}

// SendTransaction mocks the SendTransaction method
func (m *MockWalletManager) SendTransaction(ctx context.Context, chain, from, to, amount, token string) (string, error) {
	args := m.Called(ctx, chain, from, to, amount, token)
	return args.String(0), args.Error(1)
}

// EstimateGas mocks the EstimateGas method
func (m *MockWalletManager) EstimateGas(ctx context.Context, chain, from, to, amount, token string) (uint64, string, error) {
	args := m.Called(ctx, chain, from, to, amount, token)
	return args.Get(0).(uint64), args.String(1), args.Error(2)
}

// GetPendingTransactions mocks the GetPendingTransactions method
func (m *MockWalletManager) GetPendingTransactions(ctx context.Context, chain, address, transactionType string, limit, offset int) ([]*PendingTransaction, error) {
	args := m.Called(ctx, chain, address, transactionType, limit, offset)
	return args.Get(0).([]*PendingTransaction), args.Error(1)
}

// RejectTransactions mocks the RejectTransactions method
func (m *MockWalletManager) RejectTransactions(ctx context.Context, transactionIds []string, reason, details string, notifyUser, auditLog bool) ([]TransactionRejectionResult, error) {
	args := m.Called(ctx, transactionIds, reason, details, notifyUser, auditLog)
	return args.Get(0).([]TransactionRejectionResult), args.Error(1)
}

// GetTransactionHistory mocks the GetTransactionHistory method
func (m *MockWalletManager) GetTransactionHistory(ctx context.Context, address string, fromBlock, toBlock *uint64, limit, offset int) ([]*HistoricalTransaction, error) {
	args := m.Called(ctx, address, fromBlock, toBlock, limit, offset)
	return args.Get(0).([]*HistoricalTransaction), args.Error(1)
}

// GetAccounts mocks the GetAccounts method
func (m *MockWalletManager) GetAccounts(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

// AddPendingTransaction mocks the AddPendingTransaction method
func (m *MockWalletManager) AddPendingTransaction(ctx context.Context, tx *PendingTransaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

// SignMessage mocks the SignMessage method
func (m *MockWalletManager) SignMessage(ctx context.Context, address, message string) (string, error) {
	args := m.Called(ctx, address, message)
	return args.String(0), args.Error(1)
}

// UnlockWallet mocks the UnlockWallet method
func (m *MockWalletManager) UnlockWallet(password string) error {
	args := m.Called(password)
	return args.Error(0)
}

// LockWallet mocks the LockWallet method
func (m *MockWalletManager) LockWallet() {
	m.Called()
}

// IsUnlocked mocks the IsUnlocked method
func (m *MockWalletManager) IsUnlocked() bool {
	args := m.Called()
	return args.Bool(0)
}

// HasWallet mocks the HasWallet method
func (m *MockWalletManager) HasWallet() bool {
	args := m.Called()
	return args.Bool(0)
}

// GetCurrentWallet mocks the GetCurrentWallet method
func (m *MockWalletManager) GetCurrentWallet() *WalletStatus {
	args := m.Called()
	return args.Get(0).(*WalletStatus)
}