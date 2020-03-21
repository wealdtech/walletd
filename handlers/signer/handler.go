package signer

import (
	"github.com/wealdtech/walletd/services/fetcher"
	"github.com/wealdtech/walletd/services/ruler"
)

// Handler is the signer handler.
type Handler struct {
	fetcher fetcher.Service
	ruler   ruler.Service
}

// New creates a new signer handler.
func New(fetcher fetcher.Service, ruler ruler.Service) *Handler {
	return &Handler{
		fetcher: fetcher,
		ruler:   ruler,
	}
}
