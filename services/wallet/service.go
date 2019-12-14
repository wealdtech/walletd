package wallet

import (
	"github.com/wealdtech/walletd/backend"
)

// Service is the wallet service.
type Service struct {
	fetcher backend.Fetcher
}

// NewService creates a new wallet service.
func NewService(fetcher backend.Fetcher) *Service {
	return &Service{
		fetcher: fetcher,
	}
}
