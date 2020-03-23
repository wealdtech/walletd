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
			name:   "CertInfoNoAccount",
			config: []*core.CertificateInfo{{Name: "test"}},
			err:    "certificate info requires at least one account",
		},
		{
			name:   "CertInfoBadAccount",
			config: []*core.CertificateInfo{{Name: "test", Accounts: []string{""}}},
			err:    "account path cannot be blank",
		},
		{
			name:   "CertInfoBadWallet",
			config: []*core.CertificateInfo{{Name: "test", Accounts: []string{"/foo"}}},
			err:    "wallet cannot be blank",
		},
		{
			name:   "CertInfoInvalidWallet",
			config: []*core.CertificateInfo{{Name: "test", Accounts: []string{"**/foo"}}},
			err:    "invalid wallet regex **",
		},
		{
			name:   "CertInfoInvalidAccount",
			config: []*core.CertificateInfo{{Name: "test", Accounts: []string{"foo/**"}}},
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
			Name:     "client1",
			Accounts: []string{"Wallet1"},
		},
	})
	require.Nil(t, err)

	tests := []struct {
		name    string
		client  string
		account string
		result  bool
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
			name:    "Valid",
			client:  "client1",
			account: "Wallet1/valid",
			result:  true,
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := checker.Check(test.client, test.account)
			assert.Equal(t, test.result, result)
		})
	}
}
