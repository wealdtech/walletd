package storage

import "github.com/wealdtech/walletd/core"

// Service contains ability to fetch and store key-indexed state.
type Service interface {
	// FetchState fetches state for a given key.
	FetchState([]byte) (*core.State, error)
	// StoreState storess state for a given key.
	StoreState([]byte, *core.State) error
}
