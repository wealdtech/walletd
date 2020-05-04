package lua

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	e2types "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/interceptors"
	lua "github.com/yuin/gopher-lua"
)

// RunRules runs a number of rules and returns a result.
func (s *Service) RunRules(ctx context.Context,
	name string,
	wallet e2types.Wallet,
	account e2types.Account,
	populateRequestTable func(*lua.LTable) error) core.RulesResult {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ruler.lua.RunRules")
	defer span.Finish()

	pubKey := account.PublicKey().Marshal()

	accountName := fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
	log := log.WithField("account", accountName)
	storeKey := []byte(fmt.Sprintf("%s-%x", name, pubKey))
	rules := s.matchRules(ctx, name, accountName)
	now := time.Now().Unix()
	if len(rules) > 0 {
		req := &lua.LTable{}
		if populateRequestTable != nil {
			if err := populateRequestTable(req); err != nil {
				log.WithError(err).Warn("Failed to populate request table")
				return core.FAILED
			}
		}
		req.RawSetString("account", lua.LString(accountName))
		req.RawSetString("pubKey", lua.LString(fmt.Sprintf("%0x", pubKey)))
		if ip, ok := ctx.Value(&interceptors.ExternalIP{}).(string); ok {
			req.RawSetString("ip", lua.LString(ip))
		}
		if client, ok := ctx.Value(&interceptors.ClientName{}).(string); ok {
			req.RawSetString("client", lua.LString(client))
		}
		req.RawSetString("timestamp", lua.LNumber(now))

		state, err := s.store.FetchState(ctx, storeKey)
		if err != nil {
			log.WithError(err).Warn("Failed to fetch state")
			return core.FAILED
		}
		storage := &lua.LTable{}
		keys, values := state.FetchAll(ctx)

		for i := range keys {
			storage.RawSet(lua.LString(keys[i]), values[i])
		}

		for i := range rules {
			messages, result := s.runRule(ctx, rules[i], req, storage)

			// Print out any messages from the script.
			messages.ForEach(func(k lua.LValue, v lua.LValue) {
				log.WithField("rulename", rules[i].Name).Info(v)
			})

			if result == core.UNKNOWN {
				log.Warn("Unknown status from script")
				return core.FAILED
			}

			if result == core.FAILED {
				log.Warn("Script failed to complete")
				return core.FAILED
			}

			// Update state with storage.
			storage.ForEach(func(k, v lua.LValue) {
				state.Store(ctx, k.String(), v)
			})

			// Update state prior to continuing.
			err = s.store.StoreState(ctx, storeKey, state)
			if err != nil {
				log.WithError(err).Warn("Failed to update state")
				return core.FAILED
			}

			if result == core.DENIED {
				return core.DENIED
			}
		}
	}
	return core.APPROVED
}

// matchRules fetches rules that match with the request.
func (s *Service) matchRules(ctx context.Context, request string, account string) []*core.Rule {
	span, _ := opentracing.StartSpanFromContext(ctx, "ruler.lua.matchRules")
	defer span.Finish()

	res := make([]*core.Rule, 0)
	for _, rule := range s.rules {
		if rule.Matches(request, account) {
			res = append(res, rule)
		}
	}
	return res
}

func (s *Service) runRule(ctx context.Context, rule *core.Rule, req *lua.LTable, storage *lua.LTable) (*lua.LTable, core.RulesResult) {
	span, _ := opentracing.StartSpanFromContext(ctx, "ruler.lua.runRule")
	defer span.Finish()

	log := log.WithField("script", rule.Name())
	l := lua.NewState()
	defer l.Close()
	if err := l.DoString(rule.Script()); err != nil {
		log.WithError(err).Warn("Failed to parse script")
		return nil, core.FAILED
	}

	messages := &lua.LTable{}
	err := l.CallByParam(lua.P{
		Fn:      l.GetGlobal("approve"),
		NRet:    1,
		Protect: true,
	}, req, storage, messages)
	if err != nil {
		log.WithError(err).Warn("Failed to run script")
		return messages, core.FAILED
	}
	approval := l.Get(-1)
	l.Pop(1)

	switch approval.String() {
	case "Approved":
		return messages, core.APPROVED
	case "Denied":
		return messages, core.DENIED
	case "Error":
		return messages, core.FAILED
	default:
		log.WithField("approval", approval.String()).Warn("Invalid approval value returned")
		return messages, core.FAILED
	}
}
