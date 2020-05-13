// Copyright © 2020 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lua

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strconv"
	"time"

	"github.com/opentracing/opentracing-go"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/interceptors"
	lua "github.com/yuin/gopher-lua"
)

// RunRules runs a number of rules and returns a result.
func (s *Service) RunRules(ctx context.Context,
	action string,
	walletName string,
	accountName string,
	accountPubKey []byte,
	req interface{}) core.RulesResult {
	//	populateRequestTable func(*lua.LTable) error) core.RulesResult {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ruler.lua.RunRules")
	defer span.Finish()

	var lockKey [48]byte
	copy(lockKey[:], accountPubKey)
	s.locker.Lock(lockKey)
	defer s.locker.Unlock(lockKey)

	account := fmt.Sprintf("%s/%s", walletName, accountName)
	log := log.With().Str("account", account).Logger()
	rules := s.matchRules(ctx, action, account)
	now := time.Now().Unix()
	if len(rules) > 0 {
		req, err := s.populateReqData(ctx, accountName, accountPubKey, now, req)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to populate request data")
			return core.FAILED
		}

		state, err := s.fetchState(ctx, action, accountPubKey)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to fetch state")
			return core.FAILED
		}
		for i := range rules {
			messages, result := s.runRule(ctx, rules[i], req, state)

			// Print out any messages from the script.
			if messages != nil {
				messages.ForEach(func(k lua.LValue, v lua.LValue) {
					log.Info().Str("rulename", rules[i].Name()).Msg(v.String())
				})
			}

			if result == core.UNKNOWN {
				log.Warn().Msg("Unknown status from script")
				return core.FAILED
			}

			if result == core.FAILED {
				log.Warn().Msg("Script failed to complete")
				return core.FAILED
			}

			// Update state prior to continuing.
			if err := s.storeState(ctx, action, accountPubKey, state); err != nil {
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

func (s *Service) populateReqData(ctx context.Context, accountName string, pubKey []byte, now int64, req interface{}) (*lua.LTable, error) {
	reqData := &lua.LTable{}
	reqData.RawSetString("account", lua.LString(accountName))
	reqData.RawSetString("pubKey", lua.LString(fmt.Sprintf("%0x", pubKey)))
	if ip, ok := ctx.Value(&interceptors.ExternalIP{}).(string); ok {
		reqData.RawSetString("ip", lua.LString(ip))
	}
	if client, ok := ctx.Value(&interceptors.ClientName{}).(string); ok {
		reqData.RawSetString("client", lua.LString(client))
	}
	reqData.RawSetString("timestamp", lua.LNumber(now))
	switch typedReq := req.(type) {
	case *pb.ListAccountsRequest:
		s.populateListAccountsReqData(ctx, reqData, typedReq)
	case *pb.SignRequest:
		s.populateSignReqData(ctx, reqData, typedReq)
	case *pb.SignBeaconAttestationRequest:
		s.populateBeaconAttestationReqData(ctx, reqData, typedReq)
	case *pb.SignBeaconProposalRequest:
		s.populateBeaconProposalReqData(ctx, reqData, typedReq)
	}

	return reqData, nil
}

func (s *Service) populateListAccountsReqData(ctx context.Context, reqData *lua.LTable, req *pb.ListAccountsRequest) {
}

func (s *Service) populateSignReqData(ctx context.Context, reqData *lua.LTable, req *pb.SignRequest) {
	reqData.RawSetString("domain", lua.LString(fmt.Sprintf("%0x", req.Domain)))
	reqData.RawSetString("data", lua.LString(fmt.Sprintf("%0x", req.Data)))
}

func (s *Service) populateBeaconAttestationReqData(ctx context.Context, reqData *lua.LTable, req *pb.SignBeaconAttestationRequest) {
	reqData.RawSetString("domain", lua.LString(fmt.Sprintf("%0x", req.Domain)))
	reqData.RawSetString("slot", lua.LNumber(req.Data.Slot))
	reqData.RawSetString("committeeIndex", lua.LNumber(req.Data.CommitteeIndex))
	reqData.RawSetString("sourceEpoch", lua.LNumber(req.Data.Source.Epoch))
	reqData.RawSetString("sourceRoot", lua.LString(fmt.Sprintf("%0x", req.Data.Source.Root)))
	reqData.RawSetString("targetEpoch", lua.LNumber(req.Data.Target.Epoch))
	reqData.RawSetString("targetRoot", lua.LString(fmt.Sprintf("%0x", req.Data.Target.Root)))
}

func (s *Service) populateBeaconProposalReqData(ctx context.Context, reqData *lua.LTable, req *pb.SignBeaconProposalRequest) {
	reqData.RawSetString("domain", lua.LString(fmt.Sprintf("%0x", req.Domain)))
	reqData.RawSetString("slot", lua.LNumber(req.Data.Slot))
	reqData.RawSetString("proposerIndex", lua.LNumber(req.Data.ProposerIndex))
	reqData.RawSetString("bodyRoot", lua.LString(fmt.Sprintf("%0x", req.Data.BodyRoot)))
	reqData.RawSetString("parentRoot", lua.LString(fmt.Sprintf("%0x", req.Data.ParentRoot)))
	reqData.RawSetString("stateRoot", lua.LString(fmt.Sprintf("%0x", req.Data.StateRoot)))
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

type stateMap struct {
	Types  map[string]string
	Values map[string]string
}

func (s *Service) fetchState(ctx context.Context, action string, pubKey []byte) (*lua.LTable, error) {
	state := &lua.LTable{}
	key := []byte(fmt.Sprintf("%s-%x", action, pubKey))
	data, err := s.store.Fetch(ctx, key)
	if err == nil {
		inMap := &stateMap{}
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		err = dec.Decode(inMap)
		if err != nil {
			return nil, err
		}
		for k, v := range inMap.Types {
			switch v {
			case "boolean":
				switch v {
				case "true":
					state.RawSet(lua.LString(k), lua.LTrue)
				default:
					state.RawSet(lua.LString(k), lua.LFalse)
				}
			case "string":
				state.RawSet(lua.LString(k), lua.LString(inMap.Values[k]))
			case "number":
				val, err := strconv.ParseFloat(inMap.Values[k], 64)
				if err != nil {
					return nil, err
				}
				state.RawSet(lua.LString(k), lua.LNumber(val))
			default:
				log.Warn().Str("key", k).Str("value", v).Msg("Unhandled type")
			}
		}
	} else if err != core.ErrNotFound {
		return nil, err
	}
	return state, nil
}

func (s *Service) storeState(ctx context.Context, action string, pubKey []byte, state *lua.LTable) error {
	key := []byte(fmt.Sprintf("%s-%x", action, pubKey))

	outMap := &stateMap{
		Types:  make(map[string]string),
		Values: make(map[string]string),
	}
	state.ForEach(func(k lua.LValue, v lua.LValue) {
		outMap.Types[k.String()] = v.Type().String()
		outMap.Values[k.String()] = v.String()
	})
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(outMap); err != nil {
		return err
	}
	value := buf.Bytes()
	return s.store.Store(ctx, key, value)
}
