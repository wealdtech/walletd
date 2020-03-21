package locker

import (
	"sync"
)

// Service provides the features and functions for the wallet daemon.
type Service struct {
	locks        map[[48]byte]*sync.Mutex
	newLockMutex *sync.Mutex
}

// New creates a new locker service.
func New() (*Service, error) {
	return &Service{
		locks:        make(map[[48]byte]*sync.Mutex),
		newLockMutex: &sync.Mutex{},
	}, nil
}

// Lock acquires a lock for a given public key.
func (s *Service) Lock(key [48]byte) {
	lock, exists := s.locks[key]
	if !exists {
		s.newLockMutex.Lock()
		defer s.newLockMutex.Unlock()
		lock, exists = s.locks[key]
		if !exists {
			lock = &sync.Mutex{}
			s.locks[key] = lock
		}
	}
	lock.Lock()
}

// Unlock frees a lock for a given public key.
func (s *Service) Unlock(key [48]byte) {
	s.locks[key].Unlock()
}
