package walletmanager

import (
	context "context"

	empty "github.com/golang/protobuf/ptypes/empty"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Lock locks a wallet.
func (h *Handler) Lock(ctx context.Context, req *pb.LockWalletRequest) (*empty.Empty, error) {
	wallet, err := h.fetcher.FetchWallet(req.Wallet)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	wallet.Lock()
	return &empty.Empty{}, nil
}
