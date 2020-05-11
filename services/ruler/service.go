package ruler

import (
	"context"

	"github.com/wealdtech/walletd/core"
	lua "github.com/yuin/gopher-lua"
)

// Service provides an interface to run rules against a LUA engine.
type Service interface {
	RunRules(context.Context, string, string, string, []byte, func(l *lua.LTable) error) core.RulesResult
}
