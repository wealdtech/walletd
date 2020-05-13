package dummy

import (
	"context"
	"strings"
)

// DummyChecker returns true all clients and accounts except those that start with "Deny".  Only for testing.
type DummyChecker struct{}

// New creates a new dummy checker.
func New() (*DummyChecker, error) {
	return &DummyChecker{}, nil
}

// Check returns true.
func (c *DummyChecker) Check(ctx context.Context, client string, account string, operation string) bool {
	return !(strings.HasPrefix(client, "Deny") || strings.Contains(account, "/Deny"))
}
