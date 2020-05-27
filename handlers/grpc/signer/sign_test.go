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

	"github.com/stretchr/testify/require"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	hd "github.com/wealdtech/go-eth2-wallet-hd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/handlers/grpc/signer"
	"github.com/wealdtech/walletd/interceptors"
	autounlocker "github.com/wealdtech/walletd/services/autounlocker/keys"
	mockchecker "github.com/wealdtech/walletd/services/checker/mock"
	"github.com/wealdtech/walletd/services/fetcher/memfetcher"
	"github.com/wealdtech/walletd/services/locker"
	"github.com/wealdtech/walletd/services/ruler/lua"
	signersvc "github.com/wealdtech/walletd/services/signer"
	"github.com/wealdtech/walletd/services/storage/mem"
)

func TestMain(m *testing.M) {
	if err := e2types.InitBLS(); err != nil {
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestSign(t *testing.T) {
	tests := []struct {
		name    string
		client  string
		account string
		data    []byte
		domain  []byte
		state   pb.ResponseState
		err     string
	}{
		{
			name:   "Empty",
			client: "client1",
			state:  pb.ResponseState_DENIED,
		},
		{
			name:    "Good",
			client:  "client1",
			account: "Wallet 1/Account 1",
			data:    []byte("Hello, world"),
			domain:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			state:   pb.ResponseState_SUCCEEDED,
		},
	}

	handler, err := Setup()
	require.Nil(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := &pb.SignRequest{
				Id: &pb.SignRequest_Account{
					Account: test.account,
				},
				Data:   test.data,
				Domain: test.domain,
			}
			ctx := context.WithValue(context.Background(), &interceptors.ClientName{}, test.client)
			resp, err := handler.Sign(ctx, req)
			if test.err == "" {
				require.NoError(t, err)
				require.Equal(t, resp.State, test.state)
				// TODO manually calculate signature and confirm.
			} else {
				require.EqualError(t, err, test.err)
			}
		})
	}
}

// Setup sets up a test signer handler.
func Setup() (*signer.Handler, error) {
	// Create a test lister handler.
	store := scratch.New()
	encryptor := keystorev4.New()

	wallet1, err := hd.CreateWallet("Wallet 1", []byte("Wallet 1 passphrase"), store, encryptor)
	if err != nil {
		return nil, err
	}
	if err := wallet1.Unlock([]byte("Wallet 1 passphrase")); err != nil {
		return nil, err
	}
	accounts := []string{
		"Account 1",
		"Account 2",
		"Account 3",
		"Account 4",
		"A different account",
		"Deny this account",
	}
	for _, account := range accounts {
		if _, err := wallet1.CreateAccount(account, []byte(account+" passphrase")); err != nil {
			return nil, err
		}
	}
	wallet1.Lock()

	locker, err := locker.New()
	if err != nil {
		return nil, err
	}
	fetcher, err := memfetcher.New([]e2wtypes.Store{store})
	if err != nil {
		return nil, err
	}
	storage, err := mem.New()
	if err != nil {
		return nil, err
	}
	rules := make([]*core.Rule, 0)
	ruler, err := lua.New(locker, storage, rules)
	if err != nil {
		return nil, err
	}

	keysConfig := &core.KeysConfig{
		Keys: []string{"Account 1 passphrase"},
	}
	unlocker, err := autounlocker.New(context.Background(), keysConfig)
	if err != nil {
		return nil, err
	}

	checker, err := mockchecker.New()
	if err != nil {
		return nil, err
	}

	signerSvc, err := signersvc.New(unlocker, checker, fetcher, ruler)
	if err != nil {
		return nil, err
	}

	return signer.New(signerSvc), nil
}
