package walletmanager

import (
	context "context"

	empty "github.com/golang/protobuf/ptypes/empty"
	pb "github.com/wealdtech/walletd/pb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Lock locks a wallet.
func (s *Service) Lock(ctx context.Context, req *pb.LockWalletRequest) (*empty.Empty, error) {
	wallet, err := s.fetcher.FetchWallet(req.Wallet)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	wallet.Lock()
	return &empty.Empty{}, nil
}
