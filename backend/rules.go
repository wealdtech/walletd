package backend

// RulesResult represents the result of running a set of rules.
type RulesResult int

const (
	UNKNOWN RulesResult = iota
	APPROVED
	DENIED
	FAILED
)
