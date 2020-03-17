package accountmanager

import (
	context "context"

	empty "github.com/golang/protobuf/ptypes/empty"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Lock locks an account.
func (h *Handler) Lock(ctx context.Context, req *pb.LockAccountRequest) (*empty.Empty, error) {
	_, account, err := h.fetcher.FetchAccount(req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	account.Lock()
	return &empty.Empty{}, nil
}
