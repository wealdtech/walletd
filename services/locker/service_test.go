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

package locker_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/walletd/services/locker"
)

func TestLocking(t *testing.T) {
	locker, err := locker.New()
	require.Nil(t, err)

	testKey := [48]byte{}

	var wg sync.WaitGroup
	// Kick off 16 goroutines each incrementing the counter 1024 times.
	counter := 0
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 1024; i++ {
				locker.Lock(testKey)
				counter++
				locker.Unlock(testKey)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	assert.Equal(t, 16*1024, counter)
}

func TestBadUnlock(t *testing.T) {
	locker, err := locker.New()
	require.Nil(t, err)

	testKey := [48]byte{}

	assert.Panics(t, func() { locker.Unlock(testKey) })
}
