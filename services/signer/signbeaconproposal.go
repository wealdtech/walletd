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

package signer

import (
	context "context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/checker"
	"github.com/wealdtech/walletd/services/ruler"
)

// BeaconBlockHeader is a copy of the Ethereum 2 BeaconBlockHeader struct with SSZ size information.
type BeaconBlockHeader struct {
	Slot          uint64
	ProposerIndex uint64
	ParentRoot    []byte `ssz-size:"32"`
	StateRoot     []byte `ssz-size:"32"`
	BodyRoot      []byte `ssz-size:"32"`
}

// SignBeaconProposal signs a proposal for a beacon block.
func (s *Service) SignBeaconProposal(ctx context.Context, credentials *checker.Credentials, accountName string, pubKey []byte, data *ruler.SignBeaconProposalData) (core.RulesResult, []byte) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.signer.SignBeaconProposal")
	defer span.Finish()
	log := log.With().Str("action", "SignBeaconProposal").Logger()
	log.Debug().Msg("Request received")

	wallet, account, checkRes := s.preCheck(ctx, credentials, accountName, pubKey, ruler.ActionSignBeaconProposal)
	if checkRes != core.APPROVED {
		return checkRes, nil
	}
	accountName = fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
	log = log.With().Str("account", accountName).Logger()

	// Confirm approval via rules.
	result := s.ruler.RunRules(ctx, ruler.ActionSignBeaconProposal, wallet.Name(), account.Name(), account.PublicKey().Marshal(), data)
	switch result {
	case core.DENIED:
		log.Debug().Str("result", "denied").Msg("Denied by rules")
		return core.DENIED, nil
	case core.FAILED:
		log.Warn().Str("result", "failed").Msg("Rules check failed")
		return core.FAILED, nil
	}

	// Obtain the signing root of the data.
	blockHeader := &BeaconBlockHeader{
		Slot:          data.Slot,
		ProposerIndex: data.ProposerIndex,
		ParentRoot:    data.ParentRoot,
		StateRoot:     data.StateRoot,
		BodyRoot:      data.BodyRoot,
	}

	// Sign it.
	signingRoot, err := generateSigningRootFromData(ctx, blockHeader, data.Domain)
	if err != nil {
		log.Warn().Err(err).Str("result", "failed").Msg("Failed to generate signing root")
		return core.FAILED, nil
	}
	signature, err := account.Sign(signingRoot[:])
	if err != nil {
		log.Warn().Err(err).Str("result", "failed").Msg("Failed to sign")
		return core.FAILED, nil
	}

	log.Debug().Str("result", "succeeded").Msg("Success")
	return core.APPROVED, signature.Marshal()
}
