package signer

import (
	"github.com/wealdtech/walletd/backend"
)

// Service is the signer service.
type Service struct {
	fetcher backend.Fetcher
}

// NewService creates a new signer service.
func NewService(fetcher backend.Fetcher) *Service {
	return &Service{
		fetcher: fetcher,
	}
}
