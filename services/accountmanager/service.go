package accountmanager

import (
	"github.com/wealdtech/walletd/backend"
)

// Service is the account manager service.
type Service struct {
	fetcher backend.Fetcher
}

// NewService creates a new account service.
func NewService(fetcher backend.Fetcher) *Service {
	return &Service{
		fetcher: fetcher,
	}
}
