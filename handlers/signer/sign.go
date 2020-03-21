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

	wallet, account, err := h.fetchAccount(req.GetAccount(), req.GetPublicKey())
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
		res.State = pb.SignState_SUCCEEDED
	case core.DENIED:
		res.State = pb.SignState_DENIED
	case core.FAILED:
		res.State = pb.SignState_FAILED
	}

	if res.State != pb.SignState_SUCCEEDED {
		return res, nil
	}

	// Sign it.
	signingRoot, err := generateSigningRootFromRoot(req.Data, req.Domain)
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
