package accountmanager

import (
	context "context"

	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
)

// Lock locks an account.
func (h *Handler) Lock(ctx context.Context, req *pb.LockAccountRequest) (*pb.LockAccountResponse, error) {
	log.WithField("account", req.GetAccount()).Info("Lock account received")
	res := &pb.LockAccountResponse{}

	_, account, err := h.fetcher.FetchAccount(req.Account)
	if err != nil {
		log.WithError(err).WithField("result", "denied").Info("Failed to fetch account")
		res.State = pb.ResponseState_DENIED
	} else {
		account.Lock()
		log.WithField("result", "succeeded").Info("Account locked")
		res.State = pb.ResponseState_SUCCEEDED
	}
	return res, nil
}
