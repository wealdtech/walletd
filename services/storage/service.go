package storage

import (
	"context"
)

// Service contains ability to fetch and store key-indexed values.
type Service interface {
	// Fetch fetches a value for a given key.
	Fetch(context.Context, []byte) ([]byte, error)
	// Store storess a value for a given key.
	Store(context.Context, []byte, []byte) error
}
