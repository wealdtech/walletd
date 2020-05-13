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

package golang

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/interceptors"
	"github.com/wealdtech/walletd/services/ruler"
)

// RunRules runs a number of rules and returns a result.
func (s *Service) RunRules(ctx context.Context,
	action string,
	walletName string,
	accountName string,
	accountPubKey []byte,
	req interface{}) core.RulesResult {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ruler.golang.RunRules")
	defer span.Finish()

	// Do not allow multiple parallel runs of tihs code for a public key.
	var lockKey [48]byte
	copy(lockKey[:], accountPubKey)
	s.locker.Lock(lockKey)
	defer s.locker.Unlock(lockKey)

	log := log.With().Str("account", fmt.Sprintf("%s/%s", walletName, accountName)).Logger()

	metadata := s.assembleMetadata(ctx, accountName, accountPubKey)
	var result core.RulesResult
	switch action {
	case ruler.ActionSign:
		result = s.runSignRule(ctx, metadata, req.(*pb.SignRequest))
	case ruler.ActionSignBeaconProposal:
		result = s.runSignBeaconProposalRule(ctx, metadata, req.(*pb.SignBeaconProposalRequest))
	case ruler.ActionSignBeaconAttestation:
		result = s.runSignBeaconAttestationRule(ctx, metadata, req.(*pb.SignBeaconAttestationRequest))
	case ruler.ActionAccessAccount:
		result = s.runListAccountsRule(ctx, metadata, req.(*pb.ListAccountsRequest))
	}

	if result == core.UNKNOWN {
		log.Warn().Msg("Unknown result from rule")
		return core.FAILED
	}
	return result
}

// reqMetadata contains request-specific metadata.
type reqMetadata struct {
	account string
	pubKey  []byte
	ip      string
	client  string
}

func (s *Service) assembleMetadata(ctx context.Context, accountName string, pubKey []byte) *reqMetadata {
	req := &reqMetadata{
		account: accountName,
		pubKey:  pubKey,
	}

	if ip, ok := ctx.Value(&interceptors.ExternalIP{}).(string); ok {
		req.ip = ip
	}
	if client, ok := ctx.Value(&interceptors.ClientName{}).(string); ok {
		req.client = client
	}

	return req
}
