package lua

import (
	"context"
	"fmt"
	"time"

	e2types "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/backend"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/interceptors"
	lua "github.com/yuin/gopher-lua"
)

// RunRules runs a number of rules and returns a result.
func (s *Service) RunRules(ctx context.Context,
	name string,
	wallet e2types.Wallet,
	account e2types.Account,
	populateRequestTable func(*lua.LTable) error) backend.RulesResult {

	accountName := fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
	log := log.WithField("account", accountName)
	storeKey := []byte(fmt.Sprintf("%s-%x", name, account.PublicKey().Marshal()))
	rules := s.matchRules(name, accountName)
	now := time.Now().Unix()
	for i := range rules {
		log := log.WithField("script", rules[i].Name())
		l := lua.NewState()
		defer l.Close()
		if err := l.DoString(rules[i].Script()); err != nil {
			log.WithError(err).Warn("Failed to parse script")
			return backend.FAILED
		}

		req := &lua.LTable{}
		if populateRequestTable != nil {
			if err := populateRequestTable(req); err != nil {
				log.WithError(err).Warn("Failed to populate request table")
				return backend.FAILED
			}
		}
		req.RawSetString("account", lua.LString(accountName))
		req.RawSetString("pubKey", lua.LString(fmt.Sprintf("%0x", account.PublicKey().Marshal())))
		if ip, ok := ctx.Value(&interceptors.ExternalIP{}).(string); ok {
			req.RawSetString("ip", lua.LString(ip))
		}
		if client, ok := ctx.Value(&interceptors.ClientName{}).(string); ok {
			req.RawSetString("client", lua.LString(client))
		}
		req.RawSetString("timestamp", lua.LNumber(now))

		state, err := s.store.FetchState(storeKey)
		if err != nil {
			log.WithError(err).Warn("Failed to fetch state")
			return backend.FAILED
		}
		storage := l.NewTable()
		keys, values := state.FetchAll()
		for i := range keys {
			storage.RawSet(lua.LString(keys[i]), values[i])
		}

		messages := l.NewTable()

		if err := l.CallByParam(lua.P{
			Fn:      l.GetGlobal("approve"),
			NRet:    1,
			Protect: true,
		}, req, storage, messages); err != nil {
			log.WithError(err).Warn("Failed to run script")
			return backend.FAILED
		}

		approval := l.Get(-1)
		l.Pop(1)

		// Print out any messages from the script.
		messages.ForEach(func(k lua.LValue, v lua.LValue) {
			log.WithField("rule", rules[i].Name).Info(v)
		})

		// Update state with storage.
		storage.ForEach(func(k, v lua.LValue) {
			state.Store(k.String(), v)
		})

		switch approval.String() {
		case "Approved":
			// Update state prior to continuing.
			err = s.store.StoreState(storeKey, state)
			if err != nil {
				log.WithError(err).Warn("Failed to update state")
				return backend.FAILED
			}
		case "Denied":
			// Update state prior to issuing denial.
			err = s.store.StoreState(storeKey, state)
			if err != nil {
				log.WithError(err).Warn("Failed to update state")
				return backend.FAILED
			}
			return backend.DENIED
		case "Error":
			// Do not update state on a failure.
			return backend.FAILED
		default:
			// Do not update state on unknown result.
			log.WithField("result", approval.String()).Warn("Unexpected result")
			return backend.FAILED
		}
	}
	return backend.APPROVED
}

// matchRules fetches rules that match with the request.
func (s *Service) matchRules(request string, account string) []*core.Rule {
	res := make([]*core.Rule, 0)
	for _, rule := range s.rules {
		if rule.Matches(request, account) {
			res = append(res, rule)
		}
	}
	return res
}
