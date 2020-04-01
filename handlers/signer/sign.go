package signer

import (
	context "context"
	"encoding/hex"

	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
	lua "github.com/yuin/gopher-lua"
)

// Sign signs data.
func (h *Handler) Sign(ctx context.Context, req *pb.SignRequest) (*pb.SignResponse, error) {
	log.WithField("account", req.GetAccount()).WithField("pubkey", hex.EncodeToString(req.GetPublicKey())).Info("Sign request received")
	res := &pb.SignResponse{}

	if req.GetAccount() == "" && len(req.GetPublicKey()) == 0 {
		log.WithField("result", "denied").Info("Neither account nor public key supplied; denied")
		res.State = pb.ResponseState_DENIED
		return res, nil
	}

	wallet, account, err := h.fetchAccount(req.GetAccount(), req.GetPublicKey())
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
		log.WithField("result", "denied").Info("Check client access denied")
		res.State = pb.ResponseState_DENIED
		return res, nil
	}

	if !account.IsUnlocked() {
		if h.autounlocker != nil {
			unlocked, err := h.autounlocker.Unlock(ctx, wallet, account)
			if err != nil {
				res.State = pb.ResponseState_FAILED
				log.WithField("result", "failed").Info("Failed during attempt to unlock account")
				return res, nil
			}
			if !unlocked {
				res.State = pb.ResponseState_DENIED
				log.WithField("result", "denied").Info("Account is locked; signing request denied")
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
		log.WithField("result", "denied").Info("Denied by rules")
	case core.FAILED:
		log.WithField("result", "failed").Warn("Rules check failed")
		res.State = pb.ResponseState_FAILED
	}

	if res.State != pb.ResponseState_SUCCEEDED {
		return res, nil
	}

	// Sign it.
	signingRoot, err := generateSigningRootFromRoot(req.Data, req.Domain)
	if err != nil {
		log.WithError(err).WithField("result", "failed").Warn("Failed to generate signing root")
		res.State = pb.ResponseState_FAILED
		return res, nil
	}
	signature, err := account.Sign(signingRoot[:])
	if err != nil {
		log.WithError(err).WithField("result", "failed").Warn("Failed to sign")
		res.State = pb.ResponseState_FAILED
		return res, nil
	}
	res.Signature = signature.Marshal()

	log.WithField("result", "succeeded").Info("Success")
	return res, nil
}
