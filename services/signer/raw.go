package signer

import (
	context "context"

	pb "github.com/wealdtech/walletd/pb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SignRaw signs raw data.
func (s *Service) SignRaw(ctx context.Context, req *pb.SignRawRequest) (*pb.SignResponse, error) {
	account, err := s.fetcher.FetchAccount(req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !account.IsUnlocked() {
		return nil, status.Error(codes.PermissionDenied, "Account is locked")
	}

	signature, err := account.Sign(req.Data, req.Domain)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.SignResponse{Signature: signature.Marshal()}, nil
}
