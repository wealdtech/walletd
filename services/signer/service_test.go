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

package signer_test

import (
	context "context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	hd "github.com/wealdtech/go-eth2-wallet-hd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/autounlocker"
	keysunlocker "github.com/wealdtech/walletd/services/autounlocker/keys"
	"github.com/wealdtech/walletd/services/checker"
	mockchecker "github.com/wealdtech/walletd/services/checker/mock"
	"github.com/wealdtech/walletd/services/fetcher"
	"github.com/wealdtech/walletd/services/fetcher/memfetcher"
	"github.com/wealdtech/walletd/services/locker"
	"github.com/wealdtech/walletd/services/ruler"
	"github.com/wealdtech/walletd/services/ruler/lua"
	"github.com/wealdtech/walletd/services/signer"
	"github.com/wealdtech/walletd/services/storage/mem"
)

func TestMain(m *testing.M) {
	if err := e2types.InitBLS(); err != nil {
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestService(t *testing.T) {
	store := scratch.New()
	encryptor := keystorev4.New()

	wallet, err := hd.CreateWallet("Test wallet", []byte("secret"), store, encryptor)
	require.NoError(t, err)
	require.NoError(t, wallet.Unlock([]byte("secret")))

	accounts := []string{
		"Test account 1",
		"Test account 2",
	}
	for _, account := range accounts {
		_, err := wallet.CreateAccount(account, []byte(account+" passphrase"))
		require.NoError(t, err)
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

	tests := []struct {
		name     string
		unlocker autounlocker.Service
		checker  checker.Service
		fetcher  fetcher.Service
		ruler    ruler.Service
		err      string
	}{
		{
			name: "Empty",
			err:  "no unlocker provided",
		},
		{
			name:     "NoChecker",
			unlocker: unlockerSvc,
			err:      "no checker provided",
		},
		{
			name:     "NoFetcher",
			unlocker: unlockerSvc,
			checker:  checkerSvc,
			err:      "no fetcher provided",
		},
		{
			name:     "NoRuler",
			unlocker: unlockerSvc,
			checker:  checkerSvc,
			fetcher:  fetcherSvc,
			err:      "no ruler provided",
		},
		{
			name:     "Good",
			unlocker: unlockerSvc,
			checker:  checkerSvc,
			fetcher:  fetcherSvc,
			ruler:    rulerSvc,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := signer.New(test.unlocker, test.checker, test.fetcher, test.ruler)
			if test.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.err)
			}
		})
	}
}
