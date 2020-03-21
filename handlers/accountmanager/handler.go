package accountmanager

import (
	"github.com/wealdtech/walletd/services/fetcher"
	"github.com/wealdtech/walletd/services/ruler"
)

// Handler is the account manager handler.
type Handler struct {
	fetcher fetcher.Service
	ruler   ruler.Service
}

// New creates a new account manager handler.
func New(fetcher fetcher.Service, ruler ruler.Service) *Handler {
	return &Handler{
		fetcher: fetcher,
		ruler:   ruler,
	}
}
