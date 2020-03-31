package walletmanager

import (
	context "context"

	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
)

// Lock locks a wallet.
func (h *Handler) Lock(ctx context.Context, req *pb.LockWalletRequest) (*pb.LockWalletResponse, error) {
	log.WithField("wallet", req.GetWallet()).Info("Lock wallet received")
	res := &pb.LockWalletResponse{}

	wallet, err := h.fetcher.FetchWallet(req.Wallet)
	if err != nil {
		log.WithError(err).WithField("result", "denied").Info("Failed to fetch wallet")
		res.State = pb.ResponseState_DENIED
	} else {
		wallet.Lock()
		log.WithField("result", "succeeded").Info("Wallet locked")
		res.State = pb.ResponseState_SUCCEEDED
	}
	return res, nil
}
