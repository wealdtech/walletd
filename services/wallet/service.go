package wallet

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/handlers/accountmanager"
	"github.com/wealdtech/walletd/handlers/lister"
	"github.com/wealdtech/walletd/handlers/signer"
	"github.com/wealdtech/walletd/handlers/walletmanager"
	"github.com/wealdtech/walletd/interceptors"
	"github.com/wealdtech/walletd/services/fetcher/memfetcher"
	"github.com/wealdtech/walletd/services/locker"
	"github.com/wealdtech/walletd/services/ruler/lua"
	"github.com/wealdtech/walletd/services/storage/badger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Service provides the features and functions for the wallet daemon.
type Service struct {
	stores     []e2wtypes.Store
	rules      []*core.Rule
	grpcServer *grpc.Server
}

// New creates a new wallet daemon service.
func New(stores []e2wtypes.Store, rules []*core.Rule) (*Service, error) {
	return &Service{
		stores: stores,
		rules:  rules,
	}, nil
}

// ServeGRPC the wallet service over GRPC.
func (s *Service) ServeGRPC(config *core.ServerConfig) error {
	if err := s.createServer(config); err != nil {
		return err
	}

	store, err := badger.New(config.StoragePath)
	if err != nil {
		return err
	}

	locker, err := locker.New()
	if err != nil {
		return err
	}

	ruler, err := lua.New(locker, store, s.rules)
	if err != nil {
		return err
	}

	fetcher := memfetcher.New(s.stores)

	pb.RegisterWalletManagerServer(s.grpcServer, walletmanager.New(fetcher, ruler))
	pb.RegisterAccountManagerServer(s.grpcServer, accountmanager.New(fetcher, ruler))
	pb.RegisterListerServer(s.grpcServer, lister.New(fetcher, ruler))
	pb.RegisterSignerServer(s.grpcServer, signer.New(fetcher, ruler))

	err = s.Serve(config)
	if err != nil {
		return err
	}
	return nil
}

// createServer creates the GRPC server.
func (s *Service) createServer(config *core.ServerConfig) error {
	grpcOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_logrus.UnaryServerInterceptor(log.NewEntry(log.StandardLogger())),
				interceptors.SourceIPInterceptor(),
				interceptors.ClientInfoInterceptor(),
			)),
	}

	// Read in the server certificate; this is required to provide security over incoming connections.
	serverCertFile := filepath.Join(config.CertPath, fmt.Sprintf("%s.crt", config.Name))
	serverKeyFile := filepath.Join(config.CertPath, fmt.Sprintf("%s.key", config.Name))
	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		return errors.Wrap(err, "Failed to load server keypair")
	}

	// Read in the certificate authority certificate; this is required to validate client certificates on incoming connections.
	caCertFile := filepath.Join(config.CertPath, "ca.crt")
	certPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Could not read CA certificate from %s", caCertFile))
	}
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return errors.Wrap(err, "Could not add CA certificate to pool")
	}

	serverCreds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    certPool,
	})
	grpcOpts = append(grpcOpts, grpc.Creds(serverCreds))
	grpcServer := grpc.NewServer(grpcOpts...)

	s.grpcServer = grpcServer
	return nil
}

// Serve serves the GRPC server.
func (s *Service) Serve(config *core.ServerConfig) error {
	listenAddress := fmt.Sprintf(":%d", config.Port)
	conn, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}
	log.WithField("address", listenAddress).Info("Listening")

	if err := s.grpcServer.Serve(conn); err != nil {
		return fmt.Errorf("could not serve gRPC: %v", err)
	}
	return nil
}
