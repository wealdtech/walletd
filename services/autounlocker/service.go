package autounlocker

import (
	"context"

	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// Service provides an interface to automatically unlock accounts when required.
type Service interface {
	Unlock(context.Context, e2wtypes.Wallet, e2wtypes.Account) (bool, error)
}
