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

package bitcask

import (
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

// Fetch fetches the value for a given key.
func (s *Store) Fetch(key []byte) ([]byte, error) {
	value, err := s.db.Get(key)
	if err != nil {
		if err == bitcask.ErrKeyNotFound {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return value, nil
}

// Store stores the value for a given key.
func (s *Store) Store(key []byte, value []byte) error {
	if err := s.db.Put(key, value); err != nil {
		return err
	}
	return s.db.Sync()
}
