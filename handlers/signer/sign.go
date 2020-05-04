package signer

import (
	context "context"
	"encoding/hex"

	"github.com/opentracing/opentracing-go"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
	lua "github.com/yuin/gopher-lua"
)

// Sign signs data.
func (h *Handler) Sign(ctx context.Context, req *pb.SignRequest) (*pb.SignResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "handlers.signer.Sign")
	defer span.Finish()

	log := log.WithField("account", req.GetAccount()).WithField("pubkey", hex.EncodeToString(req.GetPublicKey()))
	log.Debug("Sign request received")
	res := &pb.SignResponse{}

	if req.GetAccount() == "" && len(req.GetPublicKey()) == 0 {
		log.WithField("result", "denied").Debug("Neither account nor public key supplied; denied")
		res.State = pb.ResponseState_DENIED
		return res, nil
	}

	wallet, account, err := h.fetchAccount(ctx, req.GetAccount(), req.GetPublicKey())
	if err != nil {
		log.WithError(err).WithField("result", "failed").Warn("Account unknown or inaccessible")
		res.State = pb.ResponseState_FAILED
		return res, nil
	}

	// Ensure this account is accessible by this client.
	ok, err := h.checkClientAccess(ctx, wallet, account, "Sign")
	if err != nil {
		log.WithError(err).WithField("result", "failed").Warn("Check client access failed")
		res.State = pb.ResponseState_FAILED
		return res, nil
	}
	if !ok {
		log.WithField("result", "denied").Debug("Check client access denied")
		res.State = pb.ResponseState_DENIED
		return res, nil
	}

	if !account.IsUnlocked() {
		if h.autounlocker != nil {
			unlocked, err := h.autounlocker.Unlock(ctx, wallet, account)
			if err != nil {
				log.WithField("result", "failed").Debug("Failed during attempt to unlock account")
				res.State = pb.ResponseState_FAILED
				return res, nil
			}
			if !unlocked {
				log.WithField("result", "denied").Debug("Account is locked; signing request denied")
				res.State = pb.ResponseState_DENIED
				return res, nil
			}
		}
	}

	// Confirm approval via rules.
	result := h.ruler.RunRules(
		ctx,
		"sign",
		wallet,
		account,
		func(table *lua.LTable) error {
			table.RawSetString("domain", lua.LString(hex.EncodeToString(req.Domain)))
			table.RawSetString("data", lua.LString(hex.EncodeToString(req.Data)))
			return nil
		})
	switch result {
	case core.APPROVED:
		res.State = pb.ResponseState_SUCCEEDED
	case core.DENIED:
		res.State = pb.ResponseState_DENIED
		log.WithField("result", "denied").Debug("Denied by rules")
	case core.FAILED:
		log.WithField("result", "failed").Warn("Rules check failed")
		res.State = pb.ResponseState_FAILED
	}

	if res.State != pb.ResponseState_SUCCEEDED {
		return res, nil
	}

	// Sign it.
	span, ctx = opentracing.StartSpanFromContext(ctx, "handlers.signer.Sign/Sign")
	signingRoot, err := generateSigningRootFromRoot(ctx, req.Data, req.Domain)
	if err != nil {
		log.WithError(err).WithField("result", "failed").Warn("Failed to generate signing root")
		res.State = pb.ResponseState_FAILED
		span.Finish()
		return res, nil
	}
	signature, err := account.Sign(signingRoot[:])
	if err != nil {
		log.WithError(err).WithField("result", "failed").Warn("Failed to sign")
		res.State = pb.ResponseState_FAILED
		span.Finish()
		return res, nil
	}
	res.Signature = signature.Marshal()
	span.Finish()

	log.WithField("result", "succeeded").Debug("Success")
	return res, nil
}
