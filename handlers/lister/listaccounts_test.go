package lister_test

import (
	context "context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	hd "github.com/wealdtech/go-eth2-wallet-hd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/handlers/lister"
	"github.com/wealdtech/walletd/services/checker/dummy"
	"github.com/wealdtech/walletd/services/fetcher/memfetcher"
	"github.com/wealdtech/walletd/services/locker"
	"github.com/wealdtech/walletd/services/ruler/lua"
	"github.com/wealdtech/walletd/services/storage/mem"
)

func TestMain(m *testing.M) {
	if err := e2types.InitBLS(); err != nil {
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestListAccounts(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		err      string
		accounts []string
	}{
		{
			name:  "Missing",
			paths: []string{},
		},
		{
			name:  "Empty",
			paths: []string{""},
		},
		{
			name:  "NoWallet",
			paths: []string{"/Account"},
		},
		{
			name:     "UnknownWallet",
			paths:    []string{"Unknown/.*"},
			accounts: []string{},
		},
		{
			name:     "UnknownPath",
			paths:    []string{"Wallet 1/nothinghere"},
			accounts: []string{},
		},
		{
			name:     "BadPath",
			paths:    []string{"Wallet 1/.***"},
			accounts: []string{},
		},
		{
			name:     "All",
			paths:    []string{"Wallet 1"},
			accounts: []string{"Account 1", "Account 2", "Account 3", "Account 4", "A different account"},
		},
		{
			name:     "Subset",
			paths:    []string{"Wallet 1/Account [0-9]+"},
			accounts: []string{"Account 1", "Account 2", "Account 3", "Account 4"},
		},
	}

	handler, err := Setup()
	require.Nil(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := &pb.ListAccountsRequest{Paths: test.paths}
			resp, err := handler.ListAccounts(context.Background(), req)
			if test.err == "" {
				// Result expected.
				require.Nil(t, err)
				assert.Equal(t, len(test.accounts), len(resp.Accounts))
			} else {
				// Error expected.
				require.NotNil(t, err)
				assert.Equal(t, test.err, err.Error())
			}
		})
	}
}

func Setup() (*lister.Handler, error) {
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

	checker, err := dummy.New()
	if err != nil {
		return nil, err
	}

	return lister.New(checker, fetcher, ruler), nil
}
