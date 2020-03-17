package lister

import (
	"github.com/wealdtech/walletd/backend"
	"github.com/wealdtech/walletd/services/ruler"
	"github.com/wealdtech/walletd/services/storage"
)

// Handler is the lister handler.
type Handler struct {
	fetcher backend.Fetcher
	ruler   ruler.Service
	store   storage.Service
}

// New creates a new lister handler.
func New(fetcher backend.Fetcher, ruler ruler.Service, store storage.Service) *Handler {
	return &Handler{
		fetcher: fetcher,
		ruler:   ruler,
		store:   store,
	}
}
