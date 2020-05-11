package locker

import (
	"sync"
)

// Service provides the features and functions for the wallet daemon.
type Service struct {
	//locks        map[[48]byte]*sync.Mutex
	locks        *sync.Map
	newLockMutex *sync.Mutex
}

// New creates a new locker service.
func New() (*Service, error) {
	return &Service{
		//locks:        make(map[[48]byte]*sync.Mutex),
		locks:        &sync.Map{},
		newLockMutex: &sync.Mutex{},
	}, nil
}

// Lock acquires a lock for a given public key.
func (s *Service) Lock(key [48]byte) {
	lock, exists := s.locks.Load(key)
	if !exists {
		s.newLockMutex.Lock()
		lock, exists = s.locks.Load(key)
		if !exists {
			lock = &sync.Mutex{}
			s.locks.Store(key, lock)
		}
		s.newLockMutex.Unlock()
	}
	lock.(*sync.Mutex).Lock()
}

// Unlock frees a lock for a given public key.
func (s *Service) Unlock(key [48]byte) {
	lock, exists := s.locks.Load(key)
	if !exists {
		panic("Attempt to unlock an unknown lock")
	}
	lock.(*sync.Mutex).Unlock()
}
