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

package static_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/checker/static"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name  string
		perms *core.Permissions
		err   string
	}{
		{
			name: "Nil",
			err:  "certificate info is required",
		},
		{
			name:  "NoCertInfo",
			perms: &core.Permissions{},
			err:   "certificates are required",
		},
		{
			name:  "CertListInfo",
			perms: &core.Permissions{Certs: []*core.CertificateInfo{}},
			err:   "certificate info empty",
		},
		{
			name:  "CertInfoEmpty",
			perms: &core.Permissions{Certs: []*core.CertificateInfo{{}}},
			err:   "certificate info requires a name",
		},
		{
			name:  "CertInfoNoPermissions",
			perms: &core.Permissions{Certs: []*core.CertificateInfo{{Name: "test"}}},
			err:   "certificate info requires at least one permission",
		},
		{
			name:  "CertInfoEmptyPath",
			perms: &core.Permissions{Certs: []*core.CertificateInfo{{Name: "test", Perms: []*core.CertificatePerms{{}}}}},
			err:   "invalid account path ",
		},
		{
			name:  "CertInfoBadWallet",
			perms: &core.Permissions{Certs: []*core.CertificateInfo{{Name: "test", Perms: []*core.CertificatePerms{{Path: "/foo"}}}}},
			err:   "wallet cannot be blank",
		},
		{
			name:  "CertInfoInvalidWallet",
			perms: &core.Permissions{Certs: []*core.CertificateInfo{{Name: "test", Perms: []*core.CertificatePerms{{Path: "**/foo"}}}}},
			err:   "invalid wallet regex **",
		},
		{
			name:  "CertInfoInvalidAccount",
			perms: &core.Permissions{Certs: []*core.CertificateInfo{{Name: "test", Perms: []*core.CertificatePerms{{Path: "foo/**"}}}}},
			err:   "invalid account regex **",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := static.New(context.Background(), test.perms)
			if test.err == "" {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
				require.EqualError(t, err, test.err)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	checker, err := static.New(context.Background(), &core.Permissions{
		Certs: []*core.CertificateInfo{
			{
				Name: "client1",
				Perms: []*core.CertificatePerms{
					{
						Path:       "Wallet1",
						Operations: []string{"Sign"},
					},
				},
			},
		},
	})
	require.Nil(t, err)

	tests := []struct {
		name      string
		client    string
		account   string
		operation string
		result    bool
	}{
		{
			name:    "Empty",
			client:  "",
			account: "",
			result:  false,
		},
		{
			name:    "EmptyAccount",
			client:  "client1",
			account: "",
			result:  false,
		},
		{
			name:    "WalletOnlyAccount",
			client:  "client1",
			account: "Wallet1",
			result:  false,
		},
		{
			name:    "AccountOnlyAccount",
			client:  "client1",
			account: "/valid",
			result:  false,
		},
		{
			name:    "AccountOnlyAccount",
			client:  "client1",
			account: "/valid",
			result:  false,
		},
		{
			name:    "WalletOnlyAccountTrailingSlash",
			client:  "client1",
			account: "Wallet1/",
			result:  false,
		},
		{
			name:    "UnknownClient",
			client:  "clientx",
			account: "Wallet1/valid",
			result:  false,
		},
		{
			name:    "UnknownWallet",
			client:  "client1",
			account: "Wallet2/valid",
			result:  false,
		},
		{
			name:    "MissingOperation",
			client:  "client1",
			account: "Wallet1/valid",
			result:  false,
		},
		{
			name:      "BadOperation",
			client:    "client1",
			account:   "Wallet1/valid",
			operation: "Bad",
			result:    false,
		},
		{
			name:      "Valid",
			client:    "client1",
			account:   "Wallet1/valid",
			operation: "Sign",
			result:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := checker.Check(context.Background(), test.client, test.account, test.operation)
			assert.Equal(t, test.result, result)
		})
	}
}
