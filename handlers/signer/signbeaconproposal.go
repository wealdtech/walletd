package signer

import (
	context "context"
	"encoding/hex"
	"fmt"

	"github.com/opentracing/opentracing-go"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
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
func (h *Handler) SignBeaconProposal(ctx context.Context, req *pb.SignBeaconProposalRequest) (*pb.SignResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "handlers.signer.SignBeaconProposal")
	defer span.Finish()

	log := log.With().Str("account", req.GetAccount()).Str("pubkey", hex.EncodeToString(req.GetPublicKey())).Str("action", "SignBeaconProposal").Logger()
	log.Debug().Msg("Request received")
	res := &pb.SignResponse{}

	if req.GetAccount() == "" && len(req.GetPublicKey()) == 0 {
		log.Debug().Str("result", "denied").Msg("Neither account nor public key supplied; denied")
		res.State = pb.ResponseState_DENIED
		return res, nil
	}

	wallet, account, err := h.fetchAccount(ctx, req.GetAccount(), req.GetPublicKey())
	if err != nil {
		log.Warn().Err(err).Str("result", "failed").Msg("Account unknown or inaccessible")
		res.State = pb.ResponseState_FAILED
		return res, nil
	}

	// Ensure this account is accessible by this client.
	accountName := fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
	ok, err := h.checkClientAccess(ctx, accountName, ruler.ActionSignBeaconProposal)
	if err != nil {
		log.Warn().Err(err).Str("result", "failed").Msg("Check client access failed")
		res.State = pb.ResponseState_FAILED
		return res, nil
	}
	if !ok {
		log.Debug().Str("result", "denied").Msg("Check client access denied")
		res.State = pb.ResponseState_DENIED
		return res, nil
	}

	if !account.IsUnlocked() {
		if h.autounlocker != nil {
			unlocked, err := h.autounlocker.Unlock(ctx, wallet, account)
			if err != nil {
				log.Debug().Str("result", "failed").Msg("Failed during attempt to unlock account")
				res.State = pb.ResponseState_FAILED
				return res, nil
			}
			if !unlocked {
				log.Debug().Str("result", "denied").Msg("Account is locked; signing request denied")
				res.State = pb.ResponseState_DENIED
				return res, nil
			}
		}
	}

	// Confirm approval via rules.
	result := h.ruler.RunRules(ctx, ruler.ActionSignBeaconProposal, wallet.Name(), account.Name(), account.PublicKey().Marshal(), req)
	switch result {
	case core.APPROVED:
		res.State = pb.ResponseState_SUCCEEDED
	case core.DENIED:
		log.Debug().Str("result", "denied").Msg("Denied by rules")
		res.State = pb.ResponseState_DENIED
	case core.FAILED:
		log.Warn().Str("result", "failed").Msg("Rules check failed")
		res.State = pb.ResponseState_FAILED
	}

	if res.State != pb.ResponseState_SUCCEEDED {
		return res, nil
	}

	// Create a local copy of the data; we need ssz size information to calculate the correct root.
	data := &BeaconBlockHeader{
		Slot:          req.Data.Slot,
		ProposerIndex: req.Data.ProposerIndex,
		ParentRoot:    req.Data.ParentRoot,
		StateRoot:     req.Data.StateRoot,
		BodyRoot:      req.Data.BodyRoot,
	}

	// Sign it.
	signingRoot, err := generateSigningRootFromData(ctx, data, req.Domain)
	if err != nil {
		log.Warn().Err(err).Str("result", "failed").Msg("Failed to generate signing root")
		res.State = pb.ResponseState_FAILED
		return res, nil
	}
	signature, err := account.Sign(signingRoot[:])
	if err != nil {
		log.Warn().Err(err).Str("result", "failed").Msg("Failed to sign")
		res.State = pb.ResponseState_FAILED
		return res, nil
	}
	res.Signature = signature.Marshal()

	log.Debug().Str("result", "succeeded").Msg("Success")
	return res, nil
}
