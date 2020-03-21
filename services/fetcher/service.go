package fetcher

import types "github.com/wealdtech/go-eth2-wallet-types/v2"

// Service is the interface for a wallet and account fetching service.
type Service interface {
	FetchWallet(path string) (types.Wallet, error)
	FetchAccount(path string) (types.Wallet, types.Account, error)
	FetchAccountByKey(pubKey []byte) (types.Wallet, types.Account, error)
}
