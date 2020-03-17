package storage

import "github.com/wealdtech/walletd/backend"

// Service contains ability to fetch and store key-indexed state.
type Service interface {
	// FetchState fetches state for a given key.
	FetchState([]byte) (*backend.State, error)
	// StoreState storess state for a given key.
	StoreState([]byte, *backend.State) error
}
