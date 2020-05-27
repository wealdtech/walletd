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

	"github.com/opentracing/opentracing-go"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/ruler"
)

// SignBeaconAttestation signs a attestation for a beacon block.
func (h *Handler) SignBeaconAttestation(ctx context.Context, req *pb.SignBeaconAttestationRequest) (*pb.SignResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "handlers.signer.SignBeaconAttestation")
	defer span.Finish()

	data := &ruler.SignBeaconAttestationData{
		Domain:          req.Domain,
		Slot:            req.Data.Slot,
		CommitteeIndex:  req.Data.CommitteeIndex,
		BeaconBlockRoot: req.Data.BeaconBlockRoot,
		Source: &ruler.Checkpoint{
			Epoch: req.Data.Source.Epoch,
			Root:  req.Data.Source.Root,
		},
		Target: &ruler.Checkpoint{
			Epoch: req.Data.Target.Epoch,
			Root:  req.Data.Target.Root,
		},
	}

	result, signature := h.signer.SignBeaconAttestation(ctx, h.generateCredentials(ctx), req.GetAccount(), req.GetPublicKey(), data)
	res := &pb.SignResponse{}
	switch result {
	case core.APPROVED:
		res.State = pb.ResponseState_SUCCEEDED
		res.Signature = signature
	case core.DENIED:
		res.State = pb.ResponseState_DENIED
	case core.FAILED:
		res.State = pb.ResponseState_FAILED
	default:
		res.State = pb.ResponseState_UNKNOWN
	}

	log.Debug().Str("result", "succeeded").Msg("Success")
	return res, nil
}
