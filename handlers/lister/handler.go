package lister

import (
	"github.com/wealdtech/walletd/backend"
	"github.com/wealdtech/walletd/services/ruler"
)

// Handler is the lister handler.
type Handler struct {
	fetcher backend.Fetcher
	ruler   ruler.Service
}

// New creates a new lister handler.
func New(fetcher backend.Fetcher, ruler ruler.Service) *Handler {
	return &Handler{
		fetcher: fetcher,
		ruler:   ruler,
	}
}
