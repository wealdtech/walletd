package signer

import (
	context "context"
	"encoding/hex"
	"fmt"

	"github.com/opentracing/opentracing-go"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
	lua "github.com/yuin/gopher-lua"
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
func (h *Handler) SignBeaconAttestation(ctx context.Context, req *pb.SignBeaconAttestationRequest) (*pb.SignResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "handlers.signer.SignBeaconAttestation")
	defer span.Finish()

	log := log.WithField("account", req.GetAccount()).WithField("pubkey", hex.EncodeToString(req.GetPublicKey()))
	log.Debug("Sign beacon attestation request received")
	res := &pb.SignResponse{}

	if req.GetAccount() == "" && len(req.GetPublicKey()) == 0 {
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
	ok, err := h.checkClientAccess(ctx, wallet, account, "Sign beacon attestation")
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
		log.WithField("result", "denied").Debug("Denied by rules")
		res.State = pb.ResponseState_DENIED
	case core.FAILED:
		log.WithField("result", "failed").Warn("Rules check failed")
		res.State = pb.ResponseState_FAILED
	}

	if res.State != pb.ResponseState_SUCCEEDED {
		return res, nil
	}

	// Create a local copy of the data; we need ssz size information to calculate the correct root.
	data := &BeaconAttestation{
		Slot:            req.Data.Slot,
		CommitteeIndex:  req.Data.CommitteeIndex,
		BeaconBlockRoot: req.Data.BeaconBlockRoot,
		Source: &Checkpoint{
			Epoch: req.Data.Source.Epoch,
			Root:  req.Data.Source.Root,
		},
		Target: &Checkpoint{
			Epoch: req.Data.Target.Epoch,
			Root:  req.Data.Target.Root,
		},
	}
	// Sign it.
	signingRoot, err := generateSigningRootFromData(ctx, data, req.Domain)
	if err != nil {
		log.WithError(err).WithField("result", "failed").Warn("Failed to generate signing root")
		res.State = pb.ResponseState_FAILED
		return res, nil
	}
	res.Signature, err = signRoot(ctx, account, signingRoot[:])
	if err != nil {
		log.WithError(err).WithField("result", "failed").Warn("Failed to sign")
		res.State = pb.ResponseState_FAILED
		return res, nil
	}

	log.WithField("result", "succeeded").Debug("Success")
	return res, nil
}
