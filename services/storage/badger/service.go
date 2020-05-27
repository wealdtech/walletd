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

package badger

import (
	"context"
	"encoding/gob"
	"errors"

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

	if len(key) == 0 {
		return nil, errors.New("no key provided")
	}

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

	if len(key) == 0 {
		return errors.New("no key provided")
	}

	if len(value) == 0 {
		return errors.New("no value provided")
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}
