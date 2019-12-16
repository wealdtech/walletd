package sign

import (
	context "context"

	e2types "github.com/wealdtech/go-eth2-types"
	pb "github.com/wealdtech/walletd/pb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SignAttestation signs a beacon chain validator attestation.
func (s *Service) SignAttestation(ctx context.Context, req *pb.SignAttestationRequest) (*pb.SignResponse, error) {
	account, err := s.fetcher.FetchAccount(req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !account.IsUnlocked() {
		return nil, status.Error(codes.PermissionDenied, "Account is locked")
	}

	domain := e2types.Domain(req.ForkVersion, e2types.DomainBeaconAttester)
	signature, err := account.Sign(req.Root, domain)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.SignResponse{Signature: signature.Marshal()}, nil
}
