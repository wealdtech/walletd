package backend

import (
	"github.com/wealdtech/walletd/core"
	lua "github.com/yuin/gopher-lua"
)

// RulesResult represents the result of running a set of rules.
type RulesResult int

const (
	UNKNOWN RulesResult = iota
	APPROVED
	DENIED
	FAILED
)

// RunRules runs a number of rules and returns a result.
func RunRules(
	rules []*core.Rule,
	populateRequestTable func(*lua.LState) (*lua.LTable, error),
	fetchState func() (*State, error),
	updateState func(*lua.LTable, *State) error) RulesResult {

	if len(rules) > 0 {
		for i := range rules {
			log := log.WithField("script", rules[i].Name())
			l := lua.NewState()
			defer l.Close()
			if err := l.DoString(rules[i].Script()); err != nil {
				log.WithError(err).Warn("Failed to parse script")
				return FAILED
			}

			// Request table is transient.
			luaReq, err := populateRequestTable(l)
			if err != nil {
				log.WithError(err).Warn("Failed to populate request table")
				return FAILED
			}

			state, err := fetchState()
			if err != nil {
				log.WithError(err).Warn("Failed to fetch state")
				return FAILED
			}
			luaStorage := l.NewTable()
			keys, values := state.FetchAll()
			for i := range keys {
				luaStorage.RawSet(lua.LString(keys[i]), values[i])
			}

			if err := l.CallByParam(lua.P{
				Fn:      l.GetGlobal("approve"),
				NRet:    1,
				Protect: true,
			}, luaReq, luaStorage); err != nil {
				log.WithError(err).Warn("Failed to run script")
				return FAILED
			}

			approval := l.Get(-1)
			l.Pop(1)

			switch approval.String() {
			case "Approved":
				// Update state prior to continuing.
				err = updateState(luaStorage, state)
				if err != nil {
					log.WithError(err).Warn("Failed to update state")
					return FAILED
				}
			case "Denied":
				// Update state prior to issuing denial.
				err = updateState(luaStorage, state)
				if err != nil {
					log.WithError(err).Warn("Failed to update state")
					return FAILED
				}
				return DENIED
			case "Error":
				// Do not update state on a failure.
				return FAILED
			default:
				// Do not update state on a failure.
				log.WithField("result", approval.String()).Warn("Unexpected result")
				return FAILED
			}
		}
	}
	return APPROVED
}
