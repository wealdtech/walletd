// Copyright © 2020 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memfetcher

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

// Service contains an in-memory cache of wallets and accounts.
type Service struct {
	stores        []e2wtypes.Store
	pubKeyPaths   map[[48]byte]string
	pubKeyPathsMx sync.RWMutex
	wallets       map[string]e2wtypes.Wallet
	walletsMx     sync.RWMutex
	accounts      map[string]e2wtypes.Account
	accountsMx    sync.RWMutex
}

// New creates a new in-memory fetcher.
func New(stores []e2wtypes.Store) (*Service, error) {
	if len(stores) == 0 {
		return nil, errors.New("no stores provided")
	}

	return &Service{
		stores:      stores,
		pubKeyPaths: make(map[[48]byte]string),
		wallets:     make(map[string]e2wtypes.Wallet),
		accounts:    make(map[string]e2wtypes.Account),
	}, nil
}

// FetchWallet fetches the wallet.
func (s *Service) FetchWallet(path string) (e2wtypes.Wallet, error) {
	walletName, _, err := util.WalletAndAccountNamesFromPath(path)
	if err != nil {
		return nil, err
	}

	// Return wallet from cache if present.
	s.walletsMx.RLock()
	wallet, exists := s.wallets[walletName]
	s.walletsMx.RUnlock()
	if exists {
		return wallet, nil
	}

	for _, store := range s.stores {
		wallet, err = e2w.OpenWallet(walletName, e2w.WithStore(store))
		if err == nil {
			break
		}
	}
	if wallet == nil {
		return nil, errors.New("wallet not found")
	}

	s.walletsMx.Lock()
	s.wallets[walletName] = wallet
	s.walletsMx.Unlock()
	return wallet, nil
}

// FetchAccount fetches the account given its name.
func (s *Service) FetchAccount(path string) (e2wtypes.Wallet, e2wtypes.Account, error) {
	// Fetch account and store in cache if present.
	wallet, err := s.FetchWallet(path)
	if err != nil {
		return nil, nil, err
	}

	// Return account from cache if present.
	s.accountsMx.RLock()
	account, exists := s.accounts[path]
	s.accountsMx.RUnlock()
	if exists {
		log.Debug().Str("path", path).Msg("Account found in cache; returning")
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
	s.accountsMx.Lock()
	s.accounts[path] = account
	s.accountsMx.Unlock()
	s.pubKeyPathsMx.Lock()
	s.pubKeyPaths[bytesutil.ToBytes48(account.PublicKey().Marshal())] = fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
	s.pubKeyPathsMx.Unlock()
	log.Debug().Str("path", path).Msg("Account stored in cache; returning")
	return wallet, account, nil
}

// FetchAccountByKey fetches the account given its public key.
func (s *Service) FetchAccountByKey(pubKey []byte) (e2wtypes.Wallet, e2wtypes.Account, error) {
	// See if we already know this key.
	s.pubKeyPathsMx.RLock()
	account, exists := s.pubKeyPaths[bytesutil.ToBytes48(pubKey)]
	s.pubKeyPathsMx.RUnlock()
	if exists {
		return s.FetchAccount(account)
	}

	// We don't.  Trawl wallets to find the result.
	encryptor := keystorev4.New()
	for _, store := range s.stores {
		for walletBytes := range store.RetrieveWallets() {
			wallet, err := walletFromBytes(walletBytes, store, encryptor)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to decode wallet")
				continue
			}
			for account := range wallet.Accounts() {
				if bytes.Equal(account.PublicKey().Marshal(), pubKey) {
					// Found it.
					path := fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
					s.accountsMx.Lock()
					s.accounts[path] = account
					s.accountsMx.Unlock()
					s.pubKeyPathsMx.Lock()
					s.pubKeyPaths[bytesutil.ToBytes48(pubKey)] = path
					s.pubKeyPathsMx.Unlock()
					return wallet, account, nil
				}
			}
		}
	}
	return nil, nil, errors.New("account not found")
}

func walletFromBytes(data []byte, store e2wtypes.Store, encryptor e2wtypes.Encryptor) (e2wtypes.Wallet, error) {
	if store == nil {
		return nil, errors.New("no store provided")
	}
	if encryptor == nil {
		return nil, errors.New("no encryptor provided")
	}
	if data == nil {
		return nil, errors.New("no data provided")
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
