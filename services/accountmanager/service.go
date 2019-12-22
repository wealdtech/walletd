package accountmanager

import (
	"github.com/wealdtech/walletd/backend"
)

// Service is the account manager service.
type Service struct {
	fetcher backend.Fetcher
	ruler   backend.Ruler
}

// NewService creates a new account service.
func NewService(fetcher backend.Fetcher, ruler backend.Ruler) *Service {
	return &Service{
		fetcher: fetcher,
		ruler:   ruler,
	}
}
