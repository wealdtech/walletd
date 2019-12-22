package interceptors

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// ExternalIP is a context tag for the external IP address of the request.
type ExternalIP struct{}

// SourceIPInterceptor adds the source IP address to incoming requests.
func SourceIPInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		grpcPeer, ok := peer.FromContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "Failure")
		}

		tcpAddr, ok := grpcPeer.Addr.(*net.TCPAddr)
		if !ok {
			return nil, status.Error(codes.Internal, "Failure")
		}

		ip := tcpAddr.IP.String()

		grpclog.Warningf("External IP=%s", ip)
		fmt.Printf("%s\n", ip)

		newCtx := context.WithValue(ctx, &ExternalIP{}, ip)

		return handler(newCtx, req)
	}
}
