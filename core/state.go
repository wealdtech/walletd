package core

import (
	"context"
	"sync"

	"github.com/opentracing/opentracing-go"
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
func (s *State) Store(ctx context.Context, key string, value lua.LValue) {
	span, _ := opentracing.StartSpanFromContext(ctx, "core.state.Store")
	defer span.Finish()

	s.mutex.Lock()
	s.Data[key] = value
	s.mutex.Unlock()
}

// FetchAll fetches all entries in state.
func (s *State) FetchAll(ctx context.Context) ([]string, []lua.LValue) {
	span, _ := opentracing.StartSpanFromContext(ctx, "core.state.FetchAll")
	defer span.Finish()

	s.mutex.RLock()
	keys := make([]string, len(s.Data))
	values := make([]lua.LValue, len(s.Data))
	i := 0
	for k, v := range s.Data {
		keys[i] = k
		values[i] = v
		i++
	}
	s.mutex.RUnlock()
	return keys, values
}
