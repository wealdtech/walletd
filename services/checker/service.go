package checker

// Service is the interface for checking client access to accounts.
type Service interface {
	Check(client string, account string) bool
}
