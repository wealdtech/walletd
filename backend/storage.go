package backend

import (
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// State contains key-specific state.
type State struct {
	mutex sync.RWMutex
	data  map[string]lua.LValue
}

// NewState creates a new state.
func NewState() *State {
	return &State{
		mutex: sync.RWMutex{},
		data:  make(map[string]lua.LValue),
	}
}

// Store stores an entry in state.
func (s *State) Store(key string, value lua.LValue) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[key] = value
}

// FetchAll fetches all entries in state.
func (s *State) FetchAll() ([]string, []lua.LValue) {
	s.mutex.RLock()
	keys := make([]string, len(s.data))
	values := make([]lua.LValue, len(s.data))
	i := 0
	for k, v := range s.data {
		keys[i] = k
		values[i] = v
		i++
	}
	defer s.mutex.RUnlock()
	return keys, values
}

// Storage contains storage
type Storage interface {
	// FetchBeaconProposalState fetches beacon proposal state for a given key.
	FetchBeaconProposalState([]byte) (*State, error)
	// StoreBeaconProposalState stores beacon proposal state for a given key.
	StoreBeaconProposalState([]byte, *State) error
	// FetchBeaconAttestationState fetches beacon attestation state for a given key.
	FetchBeaconAttestationState([]byte) (*State, error)
	// StoreBeaconAttestationState stores beacon attestation state for a given key.
	StoreBeaconAttestationState([]byte, *State) error
	// FetchListAccountsState fetches list accounts state for a given key.
	FetchListAccountsState([]byte) (*State, error)
	// StoreListAccountsState stores list accounts state for a given key.
	StoreListAccountsState([]byte, *State) error
}
