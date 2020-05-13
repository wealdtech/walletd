package golang

import (
	"github.com/wealdtech/walletd/services/locker"
	"github.com/wealdtech/walletd/services/storage"
)

// Service is the ruler service.
type Service struct {
	locker *locker.Service
	store  storage.Service
}

// New creates a new Go ruler service.
func New(locker *locker.Service, store storage.Service) (*Service, error) {
	return &Service{
		locker: locker,
		store:  store,
	}, nil
}
