package badger

import (
	"context"
	"encoding/gob"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"github.com/opentracing/opentracing-go"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/util"
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
	opt.Logger = util.NewLogShim(log)
	db, err := badger.Open(opt)
	if err != nil {
		return nil, err
	}

	return &Store{
		db: db,
	}, nil
}

// Fetch fetches a value for a given key.
func (s *Store) Fetch(ctx context.Context, key []byte) ([]byte, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "storage.badger.Fetch")
	defer span.Finish()

	var value []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return core.ErrNotFound
			}
		}
		err = item.Value(func(val []byte) error {
			value = val
			return nil
		})
		if err != nil {
			return err
		}

		return err
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Store stores the value for a given key.
func (s *Store) Store(ctx context.Context, key []byte, value []byte) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "storage.badger.Store")
	defer span.Finish()

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}
