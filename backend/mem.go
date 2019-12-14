package backend

import (
	"errors"
	"strings"

	e2w "github.com/wealdtech/go-eth2-wallet"
	types "github.com/wealdtech/go-eth2-wallet-types"
)

// MemFetcher contains an in-memory cache of wallets and accounts.
type MemFetcher struct {
	stores   []types.Store
	wallets  map[string]types.Wallet
	accounts map[string]types.Account
}

// NewMemFetcher creates a new in-memory fetcher.
func NewMemFetcher(stores []types.Store) Fetcher {
	return &MemFetcher{
		stores:   stores,
		wallets:  make(map[string]types.Wallet),
		accounts: make(map[string]types.Account),
	}
}

// FetchWallet fetches the wallet.
func (f *MemFetcher) FetchWallet(path string) (types.Wallet, error) {
	walletName, _, err := walletAndAccountNamesFromPath(path)
	if err != nil {
		return nil, err
	}

	// Return wallet from cache if present.
	if wallet, exists := f.wallets[walletName]; exists {
		return wallet, nil
	}

	var wallet types.Wallet
	for _, store := range f.stores {
		wallet, err = e2w.OpenWallet(walletName, e2w.WithStore(store))
		if err == nil {
			break
		}
	}
	if wallet == nil {
		return nil, errors.New("Wallet not found")
	}

	f.wallets[walletName] = wallet
	return wallet, nil
}

// FetchAccount fetches the account.
func (f *MemFetcher) FetchAccount(path string) (types.Account, error) {
	// Return account from cache if present.
	if account, exists := f.accounts[path]; exists {
		return account, nil
	}

	// Fetch account and store in cache if present.
	wallet, err := f.FetchWallet(path)
	if err != nil {
		return nil, err
	}
	_, accountName, err := walletAndAccountNamesFromPath(path)
	if err != nil {
		return nil, err
	}
	account, err := wallet.AccountByName(accountName)
	if err != nil {
		return nil, err
	}
	f.accounts[path] = account
	return account, nil
}

// walletAndAccountNamesFromPath breaks a path in to wallet and account names.
func walletAndAccountNamesFromPath(path string) (string, string, error) {
	if len(path) == 0 {
		return "", "", errors.New("invalid account format")
	}
	index := strings.Index(path, "/")
	if index == -1 {
		// Just the wallet
		return path, "", nil
	}
	if index == len(path)-1 {
		// Trailing /
		return path[:index], "", nil
	}
	return path[:index], path[index+1:], nil
}
