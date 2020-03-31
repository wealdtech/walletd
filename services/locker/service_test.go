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
