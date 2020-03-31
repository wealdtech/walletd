package badger

import (
	"bytes"
	"encoding/gob"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"github.com/wealdtech/walletd/core"
	lua "github.com/yuin/gopher-lua"
)

func init() {
	gob.Register(lua.LNumber(0))
	gob.Register(lua.LString(""))
	gob.Register(lua.LBool(false))
	gob.Register(lua.LTable{})
}

// Store holds key/value pairs in a badger database.
type Store struct {
	db *badger.DB
}

// New creates a new badger storage.
func New(base string) (*Store, error) {
	opt := badger.DefaultOptions(base)
	opt.TableLoadingMode = options.LoadToRAM
	opt.Logger = log
	db, err := badger.Open(opt)
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
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				// No key; leave the state empty.
				return nil
			}
		}
		err = item.Value(func(val []byte) error {
			buf := bytes.NewBuffer(val)
			dec := gob.NewDecoder(buf)
			return dec.Decode(&state)
		})
		if err != nil {
			return err
		}

		return err
	})
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
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}
