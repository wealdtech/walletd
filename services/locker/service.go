// Copyright © 2020 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
