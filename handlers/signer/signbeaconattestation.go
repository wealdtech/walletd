package signer

import (
	context "context"
	"encoding/hex"
	"fmt"

	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
	lua "github.com/yuin/gopher-lua"
)

// SignBeaconAttestation signs a attestation for a beacon block.
func (h *Handler) SignBeaconAttestation(ctx context.Context, req *pb.SignBeaconAttestationRequest) (*pb.SignResponse, error) {
	log.WithField("account", req.GetAccount()).WithField("pubkey", hex.EncodeToString(req.GetPublicKey())).Info("Sign beacon attestation request received")
	res := &pb.SignResponse{}

	if req.GetAccount() == "" && len(req.GetPublicKey()) == 0 {
		res.State = pb.ResponseState_DENIED
		return res, nil
	}

	wallet, account, err := h.fetchAccount(req.GetAccount(), req.GetPublicKey())
	if err != nil {
		log.WithError(err).Debug("Failed to fetch account")
		res.State = pb.ResponseState_FAILED
		return res, nil
	}

	// Ensure this account is accessible by this client.
	ok, err := h.checkClientAccess(ctx, wallet, account, "Sign beacon attestation")
	if err != nil {
		res.State = pb.ResponseState_FAILED
		return res, nil
	}
	if !ok {
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
				log.WithField("result", "denied").Debug("Account is locked; signing request denied")
				return res, nil
			}
		}
	}

	// Confirm approval via rules.
	result := h.ruler.RunRules(
		ctx,
		"sign beacon attestation",
		wallet,
		account,
		func(table *lua.LTable) error {
			table.RawSetString("domain", lua.LString(fmt.Sprintf("%0x", req.Domain)))
			table.RawSetString("slot", lua.LNumber(req.Data.Slot))
			table.RawSetString("committeeIndex", lua.LNumber(req.Data.CommitteeIndex))
			table.RawSetString("sourceEpoch", lua.LNumber(req.Data.Source.Epoch))
			table.RawSetString("sourceRoot", lua.LString(fmt.Sprintf("%0x", req.Data.Source.Root)))
			table.RawSetString("targetEpoch", lua.LNumber(req.Data.Target.Epoch))
			table.RawSetString("targetRoot", lua.LString(fmt.Sprintf("%0x", req.Data.Target.Root)))
			return nil
		})
	switch result {
	case core.APPROVED:
		res.State = pb.ResponseState_SUCCEEDED
	case core.DENIED:
		res.State = pb.ResponseState_DENIED
	case core.FAILED:
		res.State = pb.ResponseState_FAILED
	}

	if res.State != pb.ResponseState_SUCCEEDED {
		return res, nil
	}

	// Sign it.
	signingRoot, err := generateSigningRootFromData(req.Data, req.Domain)
	if err != nil {
		log.WithError(err).Warn("Failed to generate signing root")
		res.State = pb.ResponseState_FAILED
		return res, nil
	}
	signature, err := account.Sign(signingRoot[:])
	if err != nil {
		log.WithError(err).Warn("Failed to sign")
		res.State = pb.ResponseState_FAILED
		return res, nil
	}
	res.Signature = signature.Marshal()

	log.WithField("result", "succeeded").Info("Success")
	return res, nil
}
