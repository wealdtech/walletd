package walletmanager

import (
	"github.com/wealdtech/walletd/services/fetcher"
	"github.com/wealdtech/walletd/services/ruler"
)

// Handler is the wallet handler.
type Handler struct {
	fetcher fetcher.Service
	ruler   ruler.Service
}

// New creates a new wallet handler.
func New(fetcher fetcher.Service, ruler ruler.Service) *Handler {
	return &Handler{
		fetcher: fetcher,
		ruler:   ruler,
	}
}
