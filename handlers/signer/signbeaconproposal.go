package signer

import (
	context "context"
	"fmt"

	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/backend"
	lua "github.com/yuin/gopher-lua"
)

// SignBeaconProposal signs a proposal for a beacon block.
func (h *Handler) SignBeaconProposal(ctx context.Context, req *pb.SignBeaconProposalRequest) (*pb.SignResponse, error) {
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
		"sign beacon proposal",
		wallet,
		account,
		func(table *lua.LTable) error {
			table.RawSetString("domain", lua.LString(fmt.Sprintf("%0x", req.Domain)))
			table.RawSetString("slot", lua.LNumber(req.Data.Slot))
			table.RawSetString("bodyRoot", lua.LString(fmt.Sprintf("%0x", req.Data.BodyRoot)))
			table.RawSetString("parentRoot", lua.LString(fmt.Sprintf("%0x", req.Data.ParentRoot)))
			table.RawSetString("stateRoot", lua.LString(fmt.Sprintf("%0x", req.Data.StateRoot)))
			return nil
		})
	switch result {
	case backend.APPROVED:
		res.State = pb.SignState_SUCCEEDED
	case backend.DENIED:
		res.State = pb.SignState_DENIED
	case backend.FAILED:
		res.State = pb.SignState_FAILED
	}

	if res.State != pb.SignState_SUCCEEDED {
		return res, nil
	}

	// Sign it.
	signingRoot, err := generateSigningRootFromData(req.Data, req.Domain)
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
