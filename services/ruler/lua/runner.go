package lua

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/interceptors"
	lua "github.com/yuin/gopher-lua"
)

// RunRules runs a number of rules and returns a result.
func (s *Service) RunRules(ctx context.Context,
	name string,
	walletName string,
	accountName string,
	accountPubKey []byte,
	populateRequestTable func(*lua.LTable) error) core.RulesResult {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ruler.lua.RunRules")
	defer span.Finish()

	var lockKey [48]byte
	copy(lockKey[:], accountPubKey)
	s.locker.Lock(lockKey)
	defer s.locker.Unlock(lockKey)

	account := fmt.Sprintf("%s/%s", walletName, accountName)
	log := log.With().Str("account", account).Logger()
	storeKey := []byte(fmt.Sprintf("%s-%x", name, accountPubKey))
	rules := s.matchRules(ctx, name, account)
	now := time.Now().Unix()
	if len(rules) > 0 {
		req, err := s.populateReqData(ctx, accountName, accountPubKey, now, populateRequestTable)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to populate request data")
			return core.FAILED
		}

		state, err := s.store.FetchState(ctx, storeKey)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to fetch state")
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
				log.Info().Str("rulename", rules[i].Name()).Msg(v.String())
			})

			if result == core.UNKNOWN {
				log.Warn().Msg("Unknown status from script")
				return core.FAILED
			}

			if result == core.FAILED {
				log.Warn().Msg("Script failed to complete")
				return core.FAILED
			}

			// Update state with storage.
			storage.ForEach(func(k, v lua.LValue) {
				state.Store(ctx, k.String(), v)
			})

			// Update state prior to continuing.
			err = s.store.StoreState(ctx, storeKey, state)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to update state")
				return core.FAILED
			}

			if result == core.DENIED {
				return core.DENIED
			}
		}
	}
	return core.APPROVED
}

func (s *Service) populateReqData(ctx context.Context, accountName string, pubKey []byte, now int64, populateRequestTable func(*lua.LTable) error) (*lua.LTable, error) {
	req := &lua.LTable{}
	if populateRequestTable != nil {
		if err := populateRequestTable(req); err != nil {
			return nil, err
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

	return req, nil
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

	log := log.With().Str("script", rule.Name()).Logger()
	l := lua.NewState()
	defer l.Close()
	if err := l.DoString(rule.Script()); err != nil {
		log.Warn().Err(err).Msg("Failed to parse script")
		return nil, core.FAILED
	}

	messages := &lua.LTable{}
	err := l.CallByParam(lua.P{
		Fn:      l.GetGlobal("approve"),
		NRet:    1,
		Protect: true,
	}, req, storage, messages)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to run script")
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
		log.Warn().Str("approval", approval.String()).Msg("Invalid approval value returned")
		return messages, core.FAILED
	}
}
