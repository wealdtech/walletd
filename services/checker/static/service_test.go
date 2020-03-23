package static_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/checker/static"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config []*core.CertificateInfo
		err    string
	}{
		{
			name: "Nil",
			err:  "certificate info is required",
		},
		{
			name:   "NoCertInfos",
			config: []*core.CertificateInfo{},
			err:    "certificate info empty",
		},
		{
			name:   "CertInfoEmpty",
			config: []*core.CertificateInfo{{}},
			err:    "certificate info requires a name",
		},
		{
			name:   "CertInfoNoPermissions",
			config: []*core.CertificateInfo{{Name: "test"}},
			err:    "certificate info requires at least one permission",
		},
		{
			name:   "CertInfoBadPath",
			config: []*core.CertificateInfo{{Name: "test", Permissions: []*core.CertificatePermissions{{}}}},
			err:    "permission path cannot be blank",
		},
		{
			name:   "CertInfoBadWallet",
			config: []*core.CertificateInfo{{Name: "test", Permissions: []*core.CertificatePermissions{{Path: "/foo"}}}},
			err:    "wallet cannot be blank",
		},
		{
			name:   "CertInfoInvalidWallet",
			config: []*core.CertificateInfo{{Name: "test", Permissions: []*core.CertificatePermissions{{Path: "**/foo"}}}},
			err:    "invalid wallet regex **",
		},
		{
			name:   "CertInfoInvalidAccount",
			config: []*core.CertificateInfo{{Name: "test", Permissions: []*core.CertificatePermissions{{Path: "foo/**"}}}},
			err:    "invalid account regex **",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := static.New(test.config)
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
	checker, err := static.New([]*core.CertificateInfo{
		{
			Name: "client1",
			Permissions: []*core.CertificatePermissions{
				{
					Path:       "Wallet1",
					Operations: []string{"Sign"},
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
			result := checker.Check(test.client, test.account, test.operation)
			assert.Equal(t, test.result, result)
		})
	}
}
