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
	"github.com/wealdtech/walletd/services/signer"
	"github.com/wealdtech/walletd/services/storage/mem"
)

func TestSign(t *testing.T) {
	store := scratch.New()
	encryptor := keystorev4.New()

	wallet, err := hd.CreateWallet("Test wallet", []byte("secret"), store, encryptor)
	require.NoError(t, err)
	require.NoError(t, wallet.Unlock([]byte("secret")))

	accountNames := []string{
		"Test account 1",
		"Test account 2",
	}
	//	accounts := make([]e2wtypes.Account, 0)
	for _, accountName := range accountNames {
		passphrase := []byte(fmt.Sprintf("%s passphrase", accountName))
		account, err := wallet.CreateAccount(accountName, passphrase)
		require.NoError(t, err)
		require.NoError(t, account.Unlock(passphrase))
		//		accounts = append(accounts, account)
	}
	wallet.Lock()

	lockerSvc, err := locker.New()
	require.NoError(t, err)
	fetcherSvc, err := memfetcher.New([]e2wtypes.Store{store})
	require.NoError(t, err)
	storageSvc, err := mem.New()
	require.NoError(t, err)
	rules := make([]*core.Rule, 0)
	rules = append(rules, &core.Rule{})
	rulerSvc, err := lua.New(lockerSvc, storageSvc, rules)
	require.NoError(t, err)

	keysConfig := &core.KeysConfig{
		Keys: []string{"Test account 1 passphrase"},
	}
	unlockerSvc, err := keysunlocker.New(context.Background(), keysConfig)
	require.NoError(t, err)

	checkerSvc, err := mockchecker.New()
	require.NoError(t, err)

	signerSvc, err := signer.New(unlockerSvc, checkerSvc, fetcherSvc, rulerSvc)
	require.NoError(t, err)

	tests := []struct {
		name        string
		credentials *checker.Credentials
		accountName string
		pubKey      []byte
		data        *ruler.SignData
		res         core.RulesResult
	}{
		{
			name: "NoData",
			res:  core.DENIED,
		},
		{
			name: "FailPreCheck",
			data: &ruler.SignData{},
			res:  core.DENIED,
		},
		{
			name:        "Good",
			data:        &ruler.SignData{},
			credentials: &checker.Credentials{Client: "client1"},
			accountName: "Test wallet/Test account 1",
			res:         core.APPROVED,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, _ := signerSvc.Sign(context.Background(), test.credentials, test.accountName, test.pubKey, test.data)
			assert.Equal(t, test.res, res)
		})
	}
}
