package lister

import (
	"github.com/wealdtech/walletd/services/checker"
	"github.com/wealdtech/walletd/services/fetcher"
	"github.com/wealdtech/walletd/services/ruler"
)

// Handler is the lister handler.
type Handler struct {
	checker checker.Service
	fetcher fetcher.Service
	ruler   ruler.Service
}

// New creates a new lister handler.
func New(checker checker.Service, fetcher fetcher.Service, ruler ruler.Service) *Handler {
	return &Handler{
		checker: checker,
		fetcher: fetcher,
		ruler:   ruler,
	}
}
