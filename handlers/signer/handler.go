package signer

import (
	"github.com/wealdtech/walletd/backend"
	"github.com/wealdtech/walletd/services/ruler"
	"github.com/wealdtech/walletd/services/storage"
)

// Handler is the signer handler.
type Handler struct {
	fetcher backend.Fetcher
	ruler   ruler.Service
	store   storage.Service
}

// New creates a new signer handler.
func New(fetcher backend.Fetcher, ruler ruler.Service, store storage.Service) *Handler {
	return &Handler{
		fetcher: fetcher,
		ruler:   ruler,
		store:   store,
	}
}
