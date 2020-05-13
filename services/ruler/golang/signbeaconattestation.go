package golang

import (
	"context"

	"github.com/opentracing/opentracing-go"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
)

func (s *Service) runSignBeaconAttestationRule(ctx context.Context, metadata *reqMetadata, req *pb.SignBeaconAttestationRequest) core.RulesResult {
	span, _ := opentracing.StartSpanFromContext(ctx, "ruler.golang.runSignBeaconAttestationRule")
	defer span.Finish()

	return core.APPROVED
}
