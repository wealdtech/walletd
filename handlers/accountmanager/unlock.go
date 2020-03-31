package accountmanager

import (
	context "context"

	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
)

// Unlock unlocks an account.
func (h *Handler) Unlock(ctx context.Context, req *pb.UnlockAccountRequest) (*pb.UnlockAccountResponse, error) {
	log.WithField("account", req.GetAccount()).Info("Unlock account received")
	res := &pb.UnlockAccountResponse{}

	_, account, err := h.fetcher.FetchAccount(req.Account)
	if err != nil {
		log.WithError(err).WithField("result", "denied").Info("Failed to fetch account")
		res.State = pb.ResponseState_DENIED
	} else {
		err = account.Unlock(req.Passphrase)
		if err != nil {
			log.WithError(err).WithField("result", "denied").Info("Failed to unlock")
			res.State = pb.ResponseState_DENIED
		} else {
			res.State = pb.ResponseState_SUCCEEDED
		}
	}
	return res, nil
}
