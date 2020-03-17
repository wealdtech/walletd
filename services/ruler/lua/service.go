package lua

import (
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/storage"
)

// Service is the ruler service.
type Service struct {
	store storage.Service
	rules []*core.Rule
}

// New creates a new ruler service
func New(store storage.Service, rules []*core.Rule) (*Service, error) {
	return &Service{
		store: store,
		rules: rules,
	}, nil
}
