package walletmanager

import (
	"github.com/wealdtech/walletd/backend"
)

// Service is the wallet service.
type Service struct {
	fetcher backend.Fetcher
	ruler   backend.Ruler
}

// NewService creates a new wallet service.
func NewService(fetcher backend.Fetcher, ruler backend.Ruler) *Service {
	return &Service{
		fetcher: fetcher,
		ruler:   ruler,
	}
}
