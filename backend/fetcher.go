package backend

import types "github.com/wealdtech/go-eth2-wallet-types"

type Fetcher interface {
	FetchWallet(path string) (types.Wallet, error)
	FetchAccount(path string) (types.Account, error)
}
