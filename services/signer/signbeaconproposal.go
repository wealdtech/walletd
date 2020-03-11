package signer

import (
	context "context"
	"fmt"

	"github.com/wealdtech/walletd/backend"
	"github.com/wealdtech/walletd/interceptors"
	pb "github.com/wealdtech/walletd/pb/v1"
	lua "github.com/yuin/gopher-lua"
)

// SignBeaconProposal signs a proposal for a beacon block.
func (s *Service) SignBeaconProposal(ctx context.Context, req *pb.SignBeaconProposalRequest) (*pb.SignResponse, error) {
	res := &pb.SignResponse{}

	wallet, account, err := s.fetchAccount(req.GetAccount(), req.GetPublicKey())
	if err != nil {
		log.WithError(err).Debug("Failed to fetch account")
		res.State = pb.SignState_FAILED
		return res, nil
	}

	if !account.IsUnlocked() {
		log.Debug("Account is locked; signing request denied")
		res.State = pb.SignState_DENIED
		return res, nil
	}

	// Confirm approval via rules.
	accountName := fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
	rules := s.ruler.Rules("sign beacon proposal", accountName)
	result := backend.RunRules(rules,
		func(l *lua.LState) (*lua.LTable, error) {
			table := l.NewTable()
			table.RawSetString("account", lua.LString(accountName))
			table.RawSetString("pubKey", lua.LString(fmt.Sprintf("%0x", account.PublicKey().Marshal())))
			table.RawSetString("domain", lua.LString(fmt.Sprintf("%0x", req.Domain)))
			table.RawSetString("slot", lua.LNumber(req.Data.Slot))
			table.RawSetString("bodyRoot", lua.LString(fmt.Sprintf("%0x", req.Data.BodyRoot)))
			table.RawSetString("parentRoot", lua.LString(fmt.Sprintf("%0x", req.Data.ParentRoot)))
			table.RawSetString("stateRoot", lua.LString(fmt.Sprintf("%0x", req.Data.StateRoot)))
			if ip, ok := ctx.Value(&interceptors.ExternalIP{}).(string); ok {
				table.RawSetString("ip", lua.LString(ip))
			}
			return table, nil
		},
		func() (*backend.State, error) {
			return s.storage.FetchBeaconProposalState(account.PublicKey().Marshal())
		},
		func(table *lua.LTable, state *backend.State) error {
			table.ForEach(func(k, v lua.LValue) {
				state.Store(k.String(), v)
			})
			return s.storage.StoreBeaconProposalState(account.PublicKey().Marshal(), state)
		})
	switch result {
	case backend.APPROVED:
		res.State = pb.SignState_SUCCEEDED
	case backend.DENIED:
		res.State = pb.SignState_DENIED
	case backend.FAILED:
		res.State = pb.SignState_FAILED
	}

	if res.State != pb.SignState_SUCCEEDED {
		return res, nil
	}

	// Sign it.
	signingRoot, err := generateSigningRootFromData(req.Data, req.Domain)
	if err != nil {
		log.WithError(err).Warn("Failed to generate signing root")
		res.State = pb.SignState_FAILED
		return res, nil
	}
	signature, err := account.Sign(signingRoot[:])
	if err != nil {
		log.WithError(err).Warn("Failed to sign")
		res.State = pb.SignState_FAILED
		return res, nil
	}
	res.Signature = signature.Marshal()

	return res, nil
}

//// runRules runs the rules for a signature process to see if it is approved.
//func (s *Service) runRules(
//	ruleName string,
//	accountName string,
//	populateRequestTable func(*lua.LState) (*lua.LTable, error),
//	fetchState func() (*backend.State, error),
//	updateState func(*lua.LTable, *backend.State) error) pb.SignState {
//	// Work through the rules we have to follow for approval.
//	rules := s.ruler.Rules(ruleName, accountName)
//	if len(rules) > 0 {
//		for i := range rules {
//			log := log.WithField("script", rules[i].Name())
//			l := lua.NewState()
//			defer l.Close()
//			if err := l.DoString(rules[i].Script()); err != nil {
//				log.WithError(err).Warn("Failed to parse script")
//				return pb.SignState_FAILED
//			}
//
//			// Request table is transient.
//			luaReq, err := populateRequestTable(l)
//			if err != nil {
//				log.WithError(err).Warn("Failed to populate request table")
//				return pb.SignState_FAILED
//			}
//
//			state, err := fetchState()
//			if err != nil {
//				log.WithError(err).Warn("Failed to fetch state")
//				return pb.SignState_FAILED
//			}
//			luaStorage := l.NewTable()
//			keys, values := state.FetchAll()
//			for i := range keys {
//				fmt.Printf("Setting %v = %v\n", keys[i], values[i])
//				luaStorage.RawSet(lua.LString(keys[i]), values[i])
//			}
//
//			if err := l.CallByParam(lua.P{
//				Fn:      l.GetGlobal("approve"),
//				NRet:    1,
//				Protect: true,
//			}, luaReq, luaStorage); err != nil {
//				log.WithError(err).Warn("Failed to run script")
//				return pb.SignState_FAILED
//			}
//
//			approval := l.Get(-1)
//			l.Pop(1)
//
//			switch approval.String() {
//			case "Approved":
//				// Update state prior to continuing.
//				err = updateState(luaStorage, state)
//				if err != nil {
//					log.WithError(err).Warn("Failed to update state")
//					return pb.SignState_FAILED
//				}
//			case "Denied":
//				// Update state prior to issuing denial.
//				err = updateState(luaStorage, state)
//				if err != nil {
//					log.WithError(err).Warn("Failed to update state")
//					return pb.SignState_FAILED
//				}
//				return pb.SignState_DENIED
//			case "Error":
//				// Do not update state on a failure.
//				return pb.SignState_FAILED
//			default:
//				// Do not update state on a failure.
//				log.WithField("result", approval.String()).Warn("Unexpected result")
//				return pb.SignState_FAILED
//			}
//		}
//	}
//	return pb.SignState_SUCCEEDED
//}
