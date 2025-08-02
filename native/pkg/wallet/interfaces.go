package wallet

import "context"

type IWalletManager interface {
	CreateWallet(ctx context.Context, chain, password string) (address string, publicKey string, mnemonic string, err error)
	ImportWallet(ctx context.Context, mnemonic, password, chainName, derivationPath string) (address string, publicKey string, importedAt int64, err error)
	GetBalance(ctx context.Context, address string, token string) (balance string, err error)
	GetStatus(ctx context.Context) (*WalletStatus, error)
	SendTransaction(ctx context.Context, chain, from, to, amount, token string) (txHash string, err error)
	EstimateGas(ctx context.Context, chain, from, to, amount, token string) (gasLimit uint64, gasPrice string, err error)
	GetPendingTransactions(ctx context.Context, chain, address, transactionType string, limit, offset int) ([]*PendingTransaction, error)
	RejectTransactions(ctx context.Context, transactionIds []string, reason, details string, notifyUser, auditLog bool) ([]TransactionRejectionResult, error)
	GetTransactionHistory(ctx context.Context, address string, fromBlock, toBlock *uint64, limit, offset int) ([]*HistoricalTransaction, error)
	GetAccounts(ctx context.Context) ([]string, error)
	AddPendingTransaction(ctx context.Context, tx *PendingTransaction) error
	SignMessage(ctx context.Context, address, message string) (signature string, err error)
	
	// Wallet storage and security methods
	UnlockWallet(password string) error
	LockWallet()
	IsUnlocked() bool
	HasWallet() bool
	GetCurrentWallet() *WalletStatus
}
