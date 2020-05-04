package dummy

import "context"

// DummyChecker always returns true.  Only for testing.
type DummyChecker struct{}

// New creates a new dummy checker.
func New() (*DummyChecker, error) {
	return &DummyChecker{}, nil
}

// Check returns true.
func (c *DummyChecker) Check(ctx context.Context, client string, account string, operation string) bool {
	return true
}
