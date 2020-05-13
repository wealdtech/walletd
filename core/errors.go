package core

import "errors"

// ErrErrNotFound is used when an item is not found, usually when attempting to fetch it from storage.
var ErrNotFound = errors.New("not found")
