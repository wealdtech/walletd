package dummy

// DummyChecker always returns true.  Only for testing.
type DummyChecker struct{}

// New creates a new dummy checker.
func New() (*DummyChecker, error) {
	return &DummyChecker{}, nil
}

// Check returns true.
func (c *DummyChecker) Check(client string, account string, operation string) bool {
	return true
}
