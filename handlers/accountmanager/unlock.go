package accountmanager

import (
	context "context"

	empty "github.com/golang/protobuf/ptypes/empty"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Unlock unlocks an account.
func (h *Handler) Unlock(ctx context.Context, req *pb.UnlockAccountRequest) (*empty.Empty, error) {
	_, account, err := h.fetcher.FetchAccount(req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = account.Unlock(req.Passphrase)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &empty.Empty{}, nil
}
