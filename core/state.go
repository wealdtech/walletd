package core

import (
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// State contains a map of key/value pairs.
type State struct {
	mutex sync.RWMutex
	Data  map[string]lua.LValue
}

// NewState creates a new state.
func NewState() *State {
	return &State{
		mutex: sync.RWMutex{},
		Data:  make(map[string]lua.LValue),
	}
}

// Store stores an entry in state.
func (s *State) Store(key string, value lua.LValue) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Data[key] = value
}

// FetchAll fetches all entries in state.
func (s *State) FetchAll() ([]string, []lua.LValue) {
	s.mutex.RLock()
	keys := make([]string, len(s.Data))
	values := make([]lua.LValue, len(s.Data))
	i := 0
	for k, v := range s.Data {
		keys[i] = k
		values[i] = v
		i++
	}
	defer s.mutex.RUnlock()
	return keys, values
}
