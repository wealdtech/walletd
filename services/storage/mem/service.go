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

package mem

import (
	"context"
	"sync"

	"github.com/wealdtech/walletd/core"
)

// Store holds key/value pairs in-memory.
// This storage is ephemeral; it should not be used for production.
type Store struct {
	data     map[string][]byte
	statesMx sync.RWMutex
}

// New creates a new memory storage.
func New() (*Store, error) {
	return &Store{
		data: make(map[string][]byte),
	}, nil
}

// Fetch fetches a value for a given key.
func (s *Store) Fetch(ctx context.Context, key []byte) ([]byte, error) {
	s.statesMx.RLock()
	data, exists := s.data[string(key)]
	s.statesMx.RUnlock()
	if !exists {
		return nil, core.ErrNotFound
	}
	return data, nil
}

// Store stores a value for a given key.
func (s *Store) Store(ctx context.Context, key []byte, value []byte) error {
	s.statesMx.Lock()
	s.data[string(key)] = value
	s.statesMx.Unlock()
	return nil
}
