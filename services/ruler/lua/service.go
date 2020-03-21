package lua

import (
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/locker"
	"github.com/wealdtech/walletd/services/storage"
)

// Service is the ruler service.
type Service struct {
	locker *locker.Service
	store  storage.Service
	rules  []*core.Rule
}

// New creates a new ruler service
func New(locker *locker.Service, store storage.Service, rules []*core.Rule) (*Service, error) {
	return &Service{
		locker: locker,
		store:  store,
		rules:  rules,
	}, nil
}
