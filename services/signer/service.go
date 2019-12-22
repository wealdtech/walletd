package signer

import (
	"github.com/wealdtech/walletd/backend"
)

// Service is the signer service.
type Service struct {
	fetcher backend.Fetcher
	ruler   backend.Ruler
}

// NewService creates a new signer service.
func NewService(fetcher backend.Fetcher, ruler backend.Ruler) *Service {
	return &Service{
		fetcher: fetcher,
		ruler:   ruler,
	}
}
