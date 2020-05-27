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

package signer

import (
	context "context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	hd "github.com/wealdtech/go-eth2-wallet-hd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/core"
	keysunlocker "github.com/wealdtech/walletd/services/autounlocker/keys"
	"github.com/wealdtech/walletd/services/checker"
	mockchecker "github.com/wealdtech/walletd/services/checker/mock"
	"github.com/wealdtech/walletd/services/fetcher/memfetcher"
	"github.com/wealdtech/walletd/services/locker"
	"github.com/wealdtech/walletd/services/ruler"
	"github.com/wealdtech/walletd/services/ruler/lua"
	"github.com/wealdtech/walletd/services/storage/mem"
)

func TestUnlockAccount(t *testing.T) {
	store := scratch.New()
	encryptor := keystorev4.New()

	wallet, err := hd.CreateWallet("Test wallet", []byte("secret"), store, encryptor)
	require.NoError(t, err)
	require.NoError(t, wallet.Unlock([]byte("secret")))

	accountNames := []string{
		"Test account 1",
		"Test account 2",
	}
	accounts := make([]e2wtypes.Account, 0)
	for _, accountName := range accountNames {
		account, err := wallet.CreateAccount(accountName, []byte(accountName+" passphrase"))
		require.NoError(t, err)
		accounts = append(accounts, account)
	}
	wallet.Lock()

	lockerSvc, err := locker.New()
	require.NoError(t, err)
	fetcherSvc, err := memfetcher.New([]e2wtypes.Store{store})
	require.NoError(t, err)
	storageSvc, err := mem.New()
	require.NoError(t, err)
	rules := make([]*core.Rule, 0)
	rulerSvc, err := lua.New(lockerSvc, storageSvc, rules)
	require.NoError(t, err)

	keysConfig := &core.KeysConfig{
		Keys: []string{"Test account 1 passphrase"},
	}
	unlockerSvc, err := keysunlocker.New(context.Background(), keysConfig)
	require.NoError(t, err)

	checkerSvc, err := mockchecker.New()
	require.NoError(t, err)

	signerSvc, err := New(unlockerSvc, checkerSvc, fetcherSvc, rulerSvc)
	require.NoError(t, err)

	tests := []struct {
		name    string
		wallet  e2wtypes.Wallet
		account e2wtypes.Account
		res     core.RulesResult
	}{
		{
			name: "Empty",
			res:  core.DENIED,
		},
		{
			name:   "NoAccount",
			wallet: wallet,
			res:    core.DENIED,
		},
		{
			name:    "UnknownPassword",
			wallet:  wallet,
			account: accounts[1],
			res:     core.DENIED,
		},
		{
			name:    "KnownPassword",
			wallet:  wallet,
			account: accounts[0],
			res:     core.APPROVED,
		},
		{
			name:    "ReUnlock",
			wallet:  wallet,
			account: accounts[0],
			res:     core.APPROVED,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := signerSvc.unlockAccount(context.Background(), test.wallet, test.account)
			assert.Equal(t, test.res, res)
		})
	}
}

func TestCheckAccess(t *testing.T) {
	store := scratch.New()
	encryptor := keystorev4.New()

	wallet, err := hd.CreateWallet("Test wallet", []byte("secret"), store, encryptor)
	require.NoError(t, err)
	require.NoError(t, wallet.Unlock([]byte("secret")))

	accountNames := []string{
		"Test account 1",
		"Test account 2",
	}
	// accounts := make([]e2wtypes.Account, 0)
	for _, accountName := range accountNames {
		passphrase := []byte(fmt.Sprintf("%s passphrase", accountName))
		account, err := wallet.CreateAccount(accountName, passphrase)
		require.NoError(t, err)
		require.NoError(t, account.Unlock(passphrase))
		// accounts = append(accounts, account)
	}
	wallet.Lock()

	lockerSvc, err := locker.New()
	require.NoError(t, err)
	fetcherSvc, err := memfetcher.New([]e2wtypes.Store{store})
	require.NoError(t, err)
	storageSvc, err := mem.New()
	require.NoError(t, err)
	rules := make([]*core.Rule, 0)
	rulerSvc, err := lua.New(lockerSvc, storageSvc, rules)
	require.NoError(t, err)

	keysConfig := &core.KeysConfig{
		Keys: []string{"Test account 1 passphrase"},
	}
	unlockerSvc, err := keysunlocker.New(context.Background(), keysConfig)
	require.NoError(t, err)

	checkerSvc, err := mockchecker.New()
	require.NoError(t, err)

	signerSvc, err := New(unlockerSvc, checkerSvc, fetcherSvc, rulerSvc)
	require.NoError(t, err)

	tests := []struct {
		name        string
		credentials *checker.Credentials
		accountName string
		action      string
		res         core.RulesResult
	}{
		{
			name: "Empty",
			res:  core.DENIED,
		},
		{
			name:        "DeniedClient",
			credentials: &checker.Credentials{Client: "Deny"},
			accountName: "Test wallet/Test account 2",
			action:      ruler.ActionSign,
			res:         core.DENIED,
		},
		{
			name:        "Good",
			credentials: &checker.Credentials{Client: "client1"},
			accountName: "Test wallet/Test account 1",
			action:      ruler.ActionSign,
			res:         core.APPROVED,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := signerSvc.checkAccess(context.Background(), test.credentials, test.accountName, test.action)
			assert.Equal(t, test.res, res)
		})
	}
}

func TestFetchAccount(t *testing.T) {
	store := scratch.New()
	encryptor := keystorev4.New()

	wallet, err := hd.CreateWallet("Test wallet", []byte("secret"), store, encryptor)
	require.NoError(t, err)
	require.NoError(t, wallet.Unlock([]byte("secret")))

	accountNames := []string{
		"Test account 1",
		"Test account 2",
	}
	accounts := make([]e2wtypes.Account, 0)
	for _, accountName := range accountNames {
		passphrase := []byte(fmt.Sprintf("%s passphrase", accountName))
		account, err := wallet.CreateAccount(accountName, passphrase)
		require.NoError(t, err)
		require.NoError(t, account.Unlock(passphrase))
		accounts = append(accounts, account)
	}
	wallet.Lock()

	lockerSvc, err := locker.New()
	require.NoError(t, err)
	fetcherSvc, err := memfetcher.New([]e2wtypes.Store{store})
	require.NoError(t, err)
	storageSvc, err := mem.New()
	require.NoError(t, err)
	rules := make([]*core.Rule, 0)
	rulerSvc, err := lua.New(lockerSvc, storageSvc, rules)
	require.NoError(t, err)

	keysConfig := &core.KeysConfig{
		Keys: []string{"Test account 1 passphrase"},
	}
	unlockerSvc, err := keysunlocker.New(context.Background(), keysConfig)
	require.NoError(t, err)

	checkerSvc, err := mockchecker.New()
	require.NoError(t, err)

	signerSvc, err := New(unlockerSvc, checkerSvc, fetcherSvc, rulerSvc)
	require.NoError(t, err)

	tests := []struct {
		name        string
		credentials *checker.Credentials
		accountName string
		pubKey      []byte
		res         core.RulesResult
	}{
		{
			name: "Empty",
			res:  core.DENIED,
		},
		{
			name:        "UnknownAccount",
			accountName: "Unknown",
			res:         core.DENIED,
		},
		{
			name:        "KnownAccount",
			accountName: "Test wallet/Test account 1",
			res:         core.APPROVED,
		},
		{
			name:   "UnknownPubKey",
			pubKey: []byte{},
			res:    core.DENIED,
		},
		{
			name:   "KnownPubKey",
			pubKey: accounts[0].PublicKey().Marshal(),
			res:    core.APPROVED,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, res := signerSvc.fetchAccount(context.Background(), test.credentials, test.accountName, test.pubKey)
			assert.Equal(t, test.res, res)
		})
	}
}

func TestPreCheck(t *testing.T) {
	store := scratch.New()
	encryptor := keystorev4.New()

	wallet, err := hd.CreateWallet("Test wallet", []byte("secret"), store, encryptor)
	require.NoError(t, err)
	require.NoError(t, wallet.Unlock([]byte("secret")))

	accountNames := []string{
		"Test account 1",
		"Test account 2",
	}
	// accounts := make([]e2wtypes.Account, 0)
	for _, accountName := range accountNames {
		passphrase := []byte(fmt.Sprintf("%s passphrase", accountName))
		account, err := wallet.CreateAccount(accountName, passphrase)
		require.NoError(t, err)
		require.NoError(t, account.Unlock(passphrase))
		// accounts = append(accounts, account)
	}
	wallet.Lock()

	lockerSvc, err := locker.New()
	require.NoError(t, err)
	fetcherSvc, err := memfetcher.New([]e2wtypes.Store{store})
	require.NoError(t, err)
	storageSvc, err := mem.New()
	require.NoError(t, err)
	rules := make([]*core.Rule, 0)
	rulerSvc, err := lua.New(lockerSvc, storageSvc, rules)
	require.NoError(t, err)

	keysConfig := &core.KeysConfig{
		Keys: []string{"Test account 1 passphrase"},
	}
	unlockerSvc, err := keysunlocker.New(context.Background(), keysConfig)
	require.NoError(t, err)

	checkerSvc, err := mockchecker.New()
	require.NoError(t, err)

	signerSvc, err := New(unlockerSvc, checkerSvc, fetcherSvc, rulerSvc)
	require.NoError(t, err)

	tests := []struct {
		name        string
		credentials *checker.Credentials
		accountName string
		pubKey      []byte
		action      string
		res         core.RulesResult
	}{
		{
			name: "Empty",
			res:  core.DENIED,
		},
		{
			name:        "Locked",
			accountName: "Test wallet/Test account 1",
			res:         core.DENIED,
		},
		{
			name:        "Unlockable",
			credentials: &checker.Credentials{Client: "client1"},
			accountName: "Test wallet/Test account 2",
			res:         core.DENIED,
		},
		{
			name:        "Good",
			credentials: &checker.Credentials{Client: "client1"},
			accountName: "Test wallet/Test account 1",
			res:         core.APPROVED,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, res := signerSvc.preCheck(context.Background(), test.credentials, test.accountName, test.pubKey, test.action)
			assert.Equal(t, test.res, res)
		})
	}
}
