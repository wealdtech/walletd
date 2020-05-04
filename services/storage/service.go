package storage

import (
	"context"

	"github.com/wealdtech/walletd/core"
)

// Service contains ability to fetch and store key-indexed state.
type Service interface {
	// FetchState fetches state for a given key.
	FetchState(context.Context, []byte) (*core.State, error)
	// StoreState storess state for a given key.
	StoreState(context.Context, []byte, *core.State) error
}
