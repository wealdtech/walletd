package keys

import (
	"context"

	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/core"
)

// Service is an autounlocker service that holds unlock passphrases.
type Service struct {
	passphrases [][]byte
}

// New creates a new autounlocker service that holds unlock passphrases.
func New(config *core.KeysConfig) (*Service, error) {
	passphrases := make([][]byte, len(config.Keys))
	for i, key := range config.Keys {
		passphrases[i] = []byte(key)
	}

	return &Service{
		passphrases: passphrases,
	}, nil
}

// Unlock attempts to unlock an account.
func (s *Service) Unlock(ctx context.Context, wallet e2wtypes.Wallet, account e2wtypes.Account) (bool, error) {
	for _, passphrase := range s.passphrases {
		if err := account.Unlock(passphrase); err == nil {
			return true, nil
		}
	}
	return false, nil
}
