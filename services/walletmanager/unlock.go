package walletmanager

import (
	context "context"

	empty "github.com/golang/protobuf/ptypes/empty"
	pb "github.com/wealdtech/walletd/pb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Unlock unlocks a wallet.
func (s *Service) Unlock(ctx context.Context, req *pb.UnlockWalletRequest) (*empty.Empty, error) {
	wallet, err := s.fetcher.FetchWallet(req.Wallet)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = wallet.Unlock(req.Passphrase)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &empty.Empty{}, nil
}
