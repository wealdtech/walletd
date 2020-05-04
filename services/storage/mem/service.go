package mem

import (
	"context"
	"sync"

	"github.com/wealdtech/walletd/core"
)

// Store holds key/value pairs in-memory.
// This storage is ephemeral; it should not be used for production.
type Store struct {
	states   map[string]*core.State
	statesMx sync.RWMutex
}

// New creates a new memory storage.
func New() (*Store, error) {
	return &Store{
		states: make(map[string]*core.State),
	}, nil
}

// FetchState fetches state for a given key.
func (s *Store) FetchState(ctx context.Context, key []byte) (*core.State, error) {
	s.statesMx.RLock()
	state, exists := s.states[string(key)]
	s.statesMx.RUnlock()
	if !exists {
		state = core.NewState()
		s.statesMx.Lock()
		s.states[string(key)] = state
		s.statesMx.Unlock()
	}
	return state, nil
}

// StoreState stores state for a given key.
func (s *Store) StoreState(ctx context.Context, key []byte, state *core.State) error {
	s.statesMx.Lock()
	s.states[string(key)] = state
	s.statesMx.Unlock()
	return nil
}
