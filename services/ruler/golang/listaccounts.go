package golang

import (
	"context"

	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
)

func (s *Service) runListAccountsRule(ctx context.Context, metadata *reqMetadata, req *pb.ListAccountsRequest) core.RulesResult {
	return core.APPROVED
}
