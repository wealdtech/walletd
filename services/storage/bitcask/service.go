package bitcask

import (
	"bytes"
	"encoding/gob"
	"path/filepath"

	"github.com/prologic/bitcask"
	"github.com/wealdtech/walletd/core"
	lua "github.com/yuin/gopher-lua"
)

func init() {
	gob.Register(lua.LNumber(0))
	gob.Register(lua.LString(""))
	gob.Register(lua.LBool(false))
	gob.Register(lua.LTable{})
}

// Store holds key/value pairs in a bitcask database.
type Store struct {
	db *bitcask.Bitcask
}

// New creates a new bitcask storage.
func New(base string) (*Store, error) {
	db, err := bitcask.Open(filepath.Join(base, "bitcask"), bitcask.WithMaxKeySize(2048), bitcask.WithMaxValueSize(1048576))
	if err != nil {
		return nil, err
	}

	return &Store{
		db: db,
	}, nil
}

// FetchState fetches state for a given key.
func (s *Store) FetchState(key []byte) (*core.State, error) {
	state := core.NewState()
	val, err := s.db.Get(key)
	if err != nil {
		if err == bitcask.ErrKeyNotFound {
			// No key; leave the state empty.
			return state, nil
		}
		return nil, err
	}
	buf := bytes.NewBuffer(val)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// StoreState stores state for a given key.
func (s *Store) StoreState(key []byte, state *core.State) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(state); err != nil {
		return err
	}
	value := buf.Bytes()
	if err := s.db.Put(key, value); err != nil {
		return err
	}
	return s.db.Sync()
}
