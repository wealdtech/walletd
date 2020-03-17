package walletmanager

import (
	context "context"

	empty "github.com/golang/protobuf/ptypes/empty"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Unlock unlocks a wallet.
func (h *Handler) Unlock(ctx context.Context, req *pb.UnlockWalletRequest) (*empty.Empty, error) {
	wallet, err := h.fetcher.FetchWallet(req.Wallet)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = wallet.Unlock(req.Passphrase)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &empty.Empty{}, nil
}
