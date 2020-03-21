package interceptors

import (
	"context"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// ClientName is a context tag for the CN of the client's certificate.
type ClientName struct{}

// ClientInfoInterceptor adds the client certificate common name to incoming requests.
func ClientInfoInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		grpcPeer, ok := peer.FromContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "Failure")
		}

		newCtx := ctx
		authState := grpcPeer.AuthInfo.(credentials.TLSInfo).State
		if authState.HandshakeComplete {
			// TODO any further checks required here?  Validity, expiry, revocation, correct CA etc?
			peerCerts := authState.PeerCertificates
			if len(peerCerts) > 0 {
				peerCert := peerCerts[0]
				newCtx = context.WithValue(ctx, &ClientName{}, peerCert.Subject.CommonName)
				grpc_ctxtags.Extract(ctx).Set("client", peerCert.Subject.CommonName)
			}
		}
		return handler(newCtx, req)
	}
}
