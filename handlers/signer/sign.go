package signer

import (
	context "context"
	"encoding/hex"

	"github.com/opentracing/opentracing-go"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/ruler"
)

// Sign signs data.
func (h *Handler) Sign(ctx context.Context, req *pb.SignRequest) (*pb.SignResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "handlers.signer.Sign")
	defer span.Finish()

	log := log.With().Str("account", req.GetAccount()).Str("pubkey", hex.EncodeToString(req.GetPublicKey())).Str("action", "Sign").Logger()
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
	ok, err := h.checkClientAccess(ctx, wallet, account, "Sign")
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
	result := h.ruler.RunRules(ctx, ruler.ActionSign, wallet.Name(), account.Name(), account.PublicKey().Marshal(), req)
	switch result {
	case core.APPROVED:
		res.State = pb.ResponseState_SUCCEEDED
	case core.DENIED:
		res.State = pb.ResponseState_DENIED
		log.Debug().Str("result", "denied").Msg("Denied by rules")
	case core.FAILED:
		log.Warn().Str("result", "failed").Msg("Rules check failed")
		res.State = pb.ResponseState_FAILED
	}

	if res.State != pb.ResponseState_SUCCEEDED {
		return res, nil
	}

	// Sign it.
	span, ctx = opentracing.StartSpanFromContext(ctx, "handlers.signer.Sign/Sign")
	signingRoot, err := generateSigningRootFromRoot(ctx, req.Data, req.Domain)
	if err != nil {
		log.Warn().Err(err).Str("result", "failed").Msg("Failed to generate signing root")
		res.State = pb.ResponseState_FAILED
		span.Finish()
		return res, nil
	}
	signature, err := account.Sign(signingRoot[:])
	if err != nil {
		log.Warn().Err(err).Str("result", "failed").Msg("Failed to sign")
		res.State = pb.ResponseState_FAILED
		span.Finish()
		return res, nil
	}
	res.Signature = signature.Marshal()
	span.Finish()

	log.Debug().Str("result", "succeeded").Msg("Success")
	return res, nil
}
