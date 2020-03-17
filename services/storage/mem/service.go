package mem

import (
	"sync"

	"github.com/wealdtech/walletd/backend"
)

// Store holds key/value pairs in-memory.
// This storage is ephemeral; it should not be used for production.
type Store struct {
	states   map[string]*backend.State
	statesMx sync.RWMutex
}

// New creates a new memory storage.
func New() (*Store, error) {
	return &Store{
		states: make(map[string]*backend.State),
	}, nil
}

// FetchState fetches state for a given key.
func (s *Store) FetchState(key []byte) (*backend.State, error) {
	s.statesMx.RLock()
	state, exists := s.states[string(key)]
	s.statesMx.RUnlock()
	if !exists {
		state = backend.NewState()
		s.statesMx.Lock()
		s.states[string(key)] = state
		s.statesMx.Unlock()
	}
	return state, nil
}

// StoreState stores state for a given key.
func (s *Store) StoreState(key []byte, state *backend.State) error {
	s.statesMx.Lock()
	s.states[string(key)] = state
	s.statesMx.Unlock()
	return nil
}
