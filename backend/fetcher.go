package backend

import types "github.com/wealdtech/go-eth2-wallet-types/v2"

type Fetcher interface {
	FetchWallet(path string) (types.Wallet, error)
	FetchAccount(path string) (types.Wallet, types.Account, error)
	FetchAccountByKey(pubKey []byte) (types.Wallet, types.Account, error)
}
