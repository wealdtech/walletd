package wallet_test

import (
	context "context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	hd "github.com/wealdtech/go-eth2-wallet-hd"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	wtypes "github.com/wealdtech/go-eth2-wallet-types"
	"github.com/wealdtech/walletd/backend"
	pb "github.com/wealdtech/walletd/pb/v1"
	"github.com/wealdtech/walletd/services/wallet"
)

func TestListAccounts(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		err      string
		accounts []string
	}{
		{
			name: "UnknownWallet",
			path: "Unknown/.*",
			err:  "rpc error: code = NotFound desc = No such wallet",
		},
		{
			name:     "EmptyPath",
			path:     "Wallet 1/",
			accounts: []string{},
		},
		{
			name:     "All",
			path:     "Wallet 1/.*",
			accounts: []string{"Account 1", "Account 2", "Account 3", "Account 4", "A different account"},
		},
		{
			name:     "Subset",
			path:     "Wallet 1/Account [0-9]+",
			accounts: []string{"Account 1", "Account 2", "Account 3", "Account 4"},
		},
	}

	service, err := Setup()
	require.Nil(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			req := &pb.ListAccountsRequest{Path: test.path}
			resp, err := service.ListAccounts(context.Background(), req)
			if test.err == "" {
				// Result expected
				require.Nil(t, err)
				assert.Equal(t, len(test.accounts), len(resp.Accounts))
				// TODO confirm names
			} else {
				// Error expected
				require.NotNil(t, err)
				assert.Equal(t, test.err, err.Error())
			}
		})
	}
}

func Setup() (*wallet.Service, error) {
	// Create a test service
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

	fetcher := backend.NewMemFetcher([]wtypes.Store{store})

	return wallet.NewService(fetcher), nil
}
