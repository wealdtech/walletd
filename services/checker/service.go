package checker

import "context"

// Service is the interface for checking client access to accounts.
type Service interface {
	Check(ctx context.Context, client string, account string, operation string) bool
}
