package backend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/wealdtech/go-bytesutil"
	e2w "github.com/wealdtech/go-eth2-wallet"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	hd "github.com/wealdtech/go-eth2-wallet-hd/v2"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/util"
)

// MemFetcher contains an in-memory cache of wallets and accounts.
type MemFetcher struct {
	stores         []e2wtypes.Store
	pubKeyAccounts map[[48]byte]string
	wallets        map[string]e2wtypes.Wallet
	accounts       map[string]e2wtypes.Account
}

// NewMemFetcher creates a new in-memory fetcher.
func NewMemFetcher(stores []e2wtypes.Store) Fetcher {
	return &MemFetcher{
		stores:         stores,
		pubKeyAccounts: make(map[[48]byte]string),
		wallets:        make(map[string]e2wtypes.Wallet),
		accounts:       make(map[string]e2wtypes.Account),
	}
}

// FetchWallet fetches the wallet.
func (f *MemFetcher) FetchWallet(path string) (e2wtypes.Wallet, error) {
	walletName, _, err := util.WalletAndAccountNamesFromPath(path)
	if err != nil {
		return nil, err
	}

	// Return wallet from cache if present.
	if wallet, exists := f.wallets[walletName]; exists {
		return wallet, nil
	}

	var wallet e2wtypes.Wallet
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

// FetchAccount fetches the account given its name.
func (f *MemFetcher) FetchAccount(path string) (e2wtypes.Wallet, e2wtypes.Account, error) {
	// Fetch account and store in cache if present.
	wallet, err := f.FetchWallet(path)
	if err != nil {
		return nil, nil, err
	}

	// Return account from cache if present.
	if account, exists := f.accounts[path]; exists {
		return wallet, account, nil
	}

	// Need to fetch manually
	_, accountName, err := util.WalletAndAccountNamesFromPath(path)
	if err != nil {
		return nil, nil, err
	}
	account, err := wallet.AccountByName(accountName)
	if err != nil {
		return nil, nil, err
	}
	f.accounts[path] = account
	return wallet, account, nil
}

// FetchAccount fetches the account given its public key.
func (f *MemFetcher) FetchAccountByKey(pubKey []byte) (e2wtypes.Wallet, e2wtypes.Account, error) {
	// See if we already know this key.
	if account, exists := f.pubKeyAccounts[bytesutil.ToBytes48(pubKey)]; exists {
		return f.FetchAccount(account)
	}

	// We don't.  Trawl wallets to find the result.
	for _, store := range f.stores {
		encryptor := keystorev4.New()
		for walletBytes := range store.RetrieveWallets() {
			wallet, err := walletFromBytes(walletBytes, store, encryptor)
			if err != nil {
				log.WithError(err).Warn("Failed to decode wallet")
				continue
			}
			for account := range wallet.Accounts() {
				if bytes.Equal(account.PublicKey().Marshal(), pubKey) {
					// Found it.
					f.accounts[fmt.Sprintf("%s/%s", wallet.Name(), account.Name())] = account
					return wallet, account, nil
				}
			}
		}
	}
	return nil, nil, errors.New("Not implemented")
}

func walletFromBytes(data []byte, store e2wtypes.Store, encryptor e2wtypes.Encryptor) (e2wtypes.Wallet, error) {
	if store == nil {
		return nil, errors.New("no store specified")
	}
	if encryptor == nil {
		return nil, errors.New("no encryptor specified")
	}

	type walletInfo struct {
		ID   uuid.UUID `json:"uuid"`
		Name string    `json:"name"`
		Type string    `json:"type"`
	}

	info := &walletInfo{}
	err := json.Unmarshal(data, info)
	if err != nil {
		return nil, err
	}
	var wallet e2wtypes.Wallet
	switch info.Type {
	case "nd", "non-deterministic":
		wallet, err = nd.DeserializeWallet(data, store, encryptor)
	case "hd", "hierarchical deterministic":
		wallet, err = hd.DeserializeWallet(data, store, encryptor)
	default:
		return nil, fmt.Errorf("unsupported wallet type %q", info.Type)
	}
	return wallet, err
}
