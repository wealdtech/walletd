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

// Checkpoint is a copy of the Ethereum 2 Checkpoint struct with SSZ size information.
type Checkpoint struct {
	Epoch uint64
	Root  []byte `ssz-size:"32"`
}

// BeaconAttestation is a copy of the Ethereum 2 BeaconAttestation struct with SSZ size information.
type BeaconAttestation struct {
	Slot            uint64
	CommitteeIndex  uint64
	BeaconBlockRoot []byte `ssz-size:"32"`
	Source          *Checkpoint
	Target          *Checkpoint
}

// SignBeaconAttestation signs a attestation for a beacon block.
func (s *Service) SignBeaconAttestation(ctx context.Context, credentials *checker.Credentials, accountName string, pubKey []byte, data *ruler.SignBeaconAttestationData) (core.RulesResult, []byte) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "services.signer.SignBeaconAttestation")
	defer span.Finish()
	log := log.With().Str("action", "SignBeaconAttestation").Logger()
	log.Debug().Msg("Request received")

	wallet, account, checkRes := s.preCheck(ctx, credentials, accountName, pubKey, ruler.ActionSignBeaconAttestation)
	if checkRes != core.APPROVED {
		return checkRes, nil
	}
	accountName = fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
	log = log.With().Str("account", accountName).Logger()

	// Confirm approval via rules.
	result := s.ruler.RunRules(ctx, ruler.ActionSignBeaconAttestation, wallet.Name(), account.Name(), account.PublicKey().Marshal(), data)
	switch result {
	case core.DENIED:
		log.Debug().Str("result", "denied").Msg("Denied by rules")
		return core.DENIED, nil
	case core.FAILED:
		log.Warn().Str("result", "failed").Msg("Rules check failed")
		return core.FAILED, nil
	}

	// Create a local copy of the data; we need ssz size information to calculate the correct root.
	attestation := &BeaconAttestation{
		Slot:            data.Slot,
		CommitteeIndex:  data.CommitteeIndex,
		BeaconBlockRoot: data.BeaconBlockRoot,
		Source: &Checkpoint{
			Epoch: data.Source.Epoch,
			Root:  data.Source.Root,
		},
		Target: &Checkpoint{
			Epoch: data.Target.Epoch,
			Root:  data.Target.Root,
		},
	}

	// Sign it.
	signingRoot, err := generateSigningRootFromData(ctx, attestation, data.Domain)
	if err != nil {
		log.Warn().Err(err).Str("result", "failed").Msg("Failed to generate signing root")
		return core.FAILED, nil
	}
	signature, err := signRoot(ctx, account, signingRoot[:])
	if err != nil {
		log.Warn().Err(err).Str("result", "failed").Msg("Failed to sign")
		return core.FAILED, nil
	}

	log.Debug().Str("result", "succeeded").Msg("Success")
	return core.APPROVED, signature
}
