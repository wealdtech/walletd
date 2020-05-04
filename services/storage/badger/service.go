package badger

import (
	"bytes"
	"context"
	"encoding/gob"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"github.com/opentracing/opentracing-go"
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
	// Increases performance, but could result in writes being lost if a crash occurs.
	opt.SyncWrites = false
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
func (s *Store) FetchState(ctx context.Context, key []byte) (*core.State, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "storage.badger.FetchState")
	defer span.Finish()

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
func (s *Store) StoreState(ctx context.Context, key []byte, state *core.State) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "storage.badger.StoreState")
	defer span.Finish()

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
