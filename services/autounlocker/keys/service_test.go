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

package keys_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/autounlocker/keys"
	"github.com/wealdtech/walletd/testing/mock"
)

func TestUnlock(t *testing.T) {
	config := &core.KeysConfig{
		Keys: []string{"secret", "secret2"},
	}
	ctx := context.Background()
	service, err := keys.New(ctx, config)
	require.NoError(t, err)

	err = e2types.InitBLS()
	require.NoError(t, err)

	tests := []struct {
		name    string
		wallet  e2wtypes.Wallet
		account e2wtypes.Account
		err     string
		result  bool
	}{
		{
			name: "NoAccount",
			err:  "no account supplied",
		},
		{
			name:    "UnknownPassword",
			account: mock.NewAccount("Account 1", []byte("unknown secret")),
			result:  false,
		},
		{
			name:    "Good",
			account: mock.NewAccount("Account 1", []byte("secret")),
			result:  true,
		},
		{
			name:    "GoodSecondTry",
			account: mock.NewAccount("Account 1", []byte("secret2")),
			result:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := service.Unlock(ctx, test.wallet, test.account)
			if test.err != "" {
				assert.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.result, result)
			}
		})
	}
}
