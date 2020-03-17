package ruler

import (
	"context"

	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/backend"
	lua "github.com/yuin/gopher-lua"
)

// Service provides an interface to run rules against an engine.
type Service interface {
	RunRules(context.Context, string, e2wtypes.Wallet, e2wtypes.Account, func(l *lua.LTable) error) backend.RulesResult
}
