package ruler

import (
	"context"

	"github.com/wealdtech/walletd/core"
)

var (
	// ActionSign is the action of signing data.
	ActionSign = "Sign"
	// ActionSignBeaconAttestation is the action of signing a beacon attestation.
	ActionSignBeaconAttestation = "Sign beacon attestation"
	// ActionSignBeaconProposal is the action of signing a beacon proposal.
	ActionSignBeaconProposal = "Sign beacon proposal"
	// ActionAccessAccount is the action of accessing an account.
	ActionAccessAccount = "Access account"
)

// Service provides an interface to check requests against a rules engine.
type Service interface {
	// RunRules runs a set of rules for the given information.
	RunRules(context.Context, string, string, string, []byte, interface{}) core.RulesResult
}
