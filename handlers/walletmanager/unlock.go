package walletmanager

import (
	context "context"

	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
)

// Unlock unlocks a wallet.
func (h *Handler) Unlock(ctx context.Context, req *pb.UnlockWalletRequest) (*pb.UnlockWalletResponse, error) {
	log.WithField("wallet", req.GetWallet()).Info("Unlock wallet received")
	res := &pb.UnlockWalletResponse{}

	wallet, err := h.fetcher.FetchWallet(req.Wallet)
	if err != nil {
		log.WithError(err).WithField("result", "denied").Info("Failed to fetch wallet")
		res.State = pb.ResponseState_DENIED
	} else {
		err = wallet.Unlock(req.Passphrase)
		if err != nil {
			log.WithError(err).WithField("result", "denied").Info("Failed to unlock")
			res.State = pb.ResponseState_DENIED
		} else {
			res.State = pb.ResponseState_SUCCEEDED
		}
	}
	return res, nil
}
