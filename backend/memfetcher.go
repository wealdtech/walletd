package backend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

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
	stores        []e2wtypes.Store
	pubKeyPaths   map[[48]byte]string
	pubKeyPathsMx sync.RWMutex
	wallets       map[string]e2wtypes.Wallet
	walletsMx     sync.RWMutex
	accounts      map[string]e2wtypes.Account
	accountsMx    sync.RWMutex
}

// NewMemFetcher creates a new in-memory fetcher.
func NewMemFetcher(stores []e2wtypes.Store) Fetcher {
	return &MemFetcher{
		stores:      stores,
		pubKeyPaths: make(map[[48]byte]string),
		wallets:     make(map[string]e2wtypes.Wallet),
		accounts:    make(map[string]e2wtypes.Account),
	}
}

// FetchWallet fetches the wallet.
func (f *MemFetcher) FetchWallet(path string) (e2wtypes.Wallet, error) {
	walletName, _, err := util.WalletAndAccountNamesFromPath(path)
	if err != nil {
		return nil, err
	}

	// Return wallet from cache if present.
	f.walletsMx.RLock()
	wallet, exists := f.wallets[walletName]
	f.walletsMx.RUnlock()
	if exists {
		return wallet, nil
	}

	for _, store := range f.stores {
		wallet, err = e2w.OpenWallet(walletName, e2w.WithStore(store))
		if err == nil {
			break
		}
	}
	if wallet == nil {
		return nil, errors.New("Wallet not found")
	}

	f.walletsMx.Lock()
	f.wallets[walletName] = wallet
	f.walletsMx.Unlock()
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
	f.accountsMx.RLock()
	account, exists := f.accounts[path]
	f.accountsMx.RUnlock()
	if exists {
		log.WithField("path", path).Debug("Account found in cache; returning")
		return wallet, account, nil
	}

	// Need to fetch manually
	_, accountName, err := util.WalletAndAccountNamesFromPath(path)
	if err != nil {
		return nil, nil, err
	}
	account, err = wallet.AccountByName(accountName)
	if err != nil {
		return nil, nil, err
	}
	f.accountsMx.Lock()
	f.accounts[path] = account
	f.accountsMx.Unlock()
	f.pubKeyPathsMx.Lock()
	f.pubKeyPaths[bytesutil.ToBytes48(account.PublicKey().Marshal())] = fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
	f.pubKeyPathsMx.Unlock()
	log.WithField("path", path).Debug("Account stored in cache; returning")
	return wallet, account, nil
}

// FetchAccountByKey fetches the account given its public key.
func (f *MemFetcher) FetchAccountByKey(pubKey []byte) (e2wtypes.Wallet, e2wtypes.Account, error) {
	// See if we already know this key.
	f.pubKeyPathsMx.RLock()
	account, exists := f.pubKeyPaths[bytesutil.ToBytes48(pubKey)]
	f.pubKeyPathsMx.RUnlock()
	if exists {
		return f.FetchAccount(account)
	}

	// We don't.  Trawl wallets to find the result.
	encryptor := keystorev4.New()
	for _, store := range f.stores {
		for walletBytes := range store.RetrieveWallets() {
			wallet, err := walletFromBytes(walletBytes, store, encryptor)
			if err != nil {
				log.WithError(err).Warn("Failed to decode wallet")
				continue
			}
			for account := range wallet.Accounts() {
				if bytes.Equal(account.PublicKey().Marshal(), pubKey) {
					// Found it.
					path := fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
					f.accountsMx.Lock()
					f.accounts[path] = account
					f.accountsMx.Unlock()
					f.pubKeyPathsMx.Lock()
					f.pubKeyPaths[bytesutil.ToBytes48(pubKey)] = path
					f.pubKeyPathsMx.Unlock()
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
