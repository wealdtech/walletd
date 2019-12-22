package signer

import (
	context "context"
	"encoding/hex"
	"fmt"

	"github.com/wealdtech/walletd/interceptors"
	pb "github.com/wealdtech/walletd/pb/v1"
	lua "github.com/yuin/gopher-lua"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Sign signs data.
func (s *Service) Sign(ctx context.Context, req *pb.SignRequest) (*pb.SignResponse, error) {
	account, err := s.fetcher.FetchAccount(req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !account.IsUnlocked() {
		return nil, status.Error(codes.PermissionDenied, "Account is locked")
	}

	// Work through the rules we have to follow.
	rules := s.ruler.Rules("sign", req.Account)
	if len(rules) > 0 {
		for i := range rules {
			l := lua.NewState()
			defer l.Close()
			if err := l.DoString(rules[i].Script()); err != nil {
				fmt.Printf("DoString() failed: %v\n", err)
				return nil, status.Error(codes.Internal, "Signing request denied")
			}

			luaReq := l.NewTable()
			luaReq.RawSetString("account", lua.LString(req.Account))
			luaReq.RawSetString("domain", lua.LNumber(float64(req.Domain)))
			luaReq.RawSetString("data", lua.LString(hex.EncodeToString(req.Data)))
			if ip, ok := ctx.Value(&interceptors.ExternalIP{}).(string); ok {
				luaReq.RawSetString("ip", lua.LString(ip))
			}

			if err := l.CallByParam(lua.P{
				Fn:      l.GetGlobal("approve"),
				NRet:    1,
				Protect: true,
			}, luaReq); err != nil {
				fmt.Printf("CallByParam() failed: %v\n", err)
				return nil, status.Error(codes.Internal, "Signing request denied")
			}

			approval := l.Get(-1)
			l.Pop(1)
			fmt.Printf("Approval: %s\n", approval.String())
		}
	}

	signature, err := account.Sign(req.Data, req.Domain)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.SignResponse{Signature: signature.Marshal()}, nil
}
