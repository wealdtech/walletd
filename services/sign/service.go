package sign

import (
	"github.com/wealdtech/walletd/backend"
)

// Service is the signing service.
type Service struct {
	fetcher backend.Fetcher
}

// NewService creates a new signing service.
func NewService(fetcher backend.Fetcher) *Service {
	return &Service{
		fetcher: fetcher,
	}
}
