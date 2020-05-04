package keys

import (
	"context"

	"github.com/opentracing/opentracing-go"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/core"
)

// Service is an autounlocker service that holds unlock passphrases.
type Service struct {
	passphrases [][]byte
}

// New creates a new autounlocker service that holds unlock passphrases.
func New(ctx context.Context, config *core.KeysConfig) (*Service, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "autounlocker.keys.New")
	defer span.Finish()

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
	span, _ := opentracing.StartSpanFromContext(ctx, "autounlocker.keys.Unlock")
	defer span.Finish()
	for _, passphrase := range s.passphrases {
		if err := account.Unlock(passphrase); err == nil {
			return true, nil
		}
	}
	return false, nil
}
