package wallet

import "context"

type IWalletManager interface {
	CreateWallet(ctx context.Context, chain string) (address string, publicKey string, err error)
	ImportWallet(ctx context.Context, mnemonic, password, chainName, derivationPath string) (address string, publicKey string, importedAt int64, err error)
	GetBalance(ctx context.Context, address string, token string) (balance string, err error)
	GetStatus(ctx context.Context) (*WalletStatus, error)
	SendTransaction(ctx context.Context, chain, from, to, amount, token string) (txHash string, err error)
	EstimateGas(ctx context.Context, chain, from, to, amount, token string) (gasLimit uint64, gasPrice string, err error)
}
