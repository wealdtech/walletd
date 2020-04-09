package signer

import (
	context "context"
	"encoding/hex"
	"fmt"

	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
	lua "github.com/yuin/gopher-lua"
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
	log := log.WithField("account", req.GetAccount()).WithField("pubkey", hex.EncodeToString(req.GetPublicKey()))
	log.Info("Sign beacon proposal request received")
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
	ok, err := h.checkClientAccess(ctx, wallet, account, "Sign beacon proposal")
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
				log.WithField("result", "failed").Info("Failed during attempt to unlock account")
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
		"sign beacon proposal",
		wallet,
		account,
		func(table *lua.LTable) error {
			table.RawSetString("domain", lua.LString(fmt.Sprintf("%0x", req.Domain)))
			table.RawSetString("slot", lua.LNumber(req.Data.Slot))
			table.RawSetString("proposerIndex", lua.LNumber(req.Data.ProposerIndex))
			table.RawSetString("bodyRoot", lua.LString(fmt.Sprintf("%0x", req.Data.BodyRoot)))
			table.RawSetString("parentRoot", lua.LString(fmt.Sprintf("%0x", req.Data.ParentRoot)))
			table.RawSetString("stateRoot", lua.LString(fmt.Sprintf("%0x", req.Data.StateRoot)))
			return nil
		})
	switch result {
	case core.APPROVED:
		res.State = pb.ResponseState_SUCCEEDED
	case core.DENIED:
		log.WithField("result", "denied").Info("Denied by rules")
		res.State = pb.ResponseState_DENIED
	case core.FAILED:
		log.WithField("result", "failed").Warn("Rules check failed")
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
	signingRoot, err := generateSigningRootFromData(data, req.Domain)
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
