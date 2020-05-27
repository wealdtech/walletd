// Copyright © 2020 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wallet

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/handlers/grpc/accountmanager"
	"github.com/wealdtech/walletd/handlers/grpc/lister"
	signerhandler "github.com/wealdtech/walletd/handlers/grpc/signer"
	"github.com/wealdtech/walletd/handlers/grpc/walletmanager"
	"github.com/wealdtech/walletd/interceptors"
	"github.com/wealdtech/walletd/services/autounlocker"
	"github.com/wealdtech/walletd/services/checker"
	"github.com/wealdtech/walletd/services/fetcher/memfetcher"
	"github.com/wealdtech/walletd/services/locker"
	"github.com/wealdtech/walletd/services/ruler"
	"github.com/wealdtech/walletd/services/ruler/golang"
	"github.com/wealdtech/walletd/services/ruler/lua"
	signersvc "github.com/wealdtech/walletd/services/signer"
	"github.com/wealdtech/walletd/services/storage/badger"
	"github.com/wealdtech/walletd/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

// Service provides the features and functions for the wallet daemon.
type Service struct {
	autounlocker autounlocker.Service
	checker      checker.Service
	stores       []e2wtypes.Store
	rules        []*core.Rule
	grpcServer   *grpc.Server
}

// New creates a new wallet daemon service.
func New(ctx context.Context, autounlocker autounlocker.Service, checker checker.Service, stores []e2wtypes.Store, rules []*core.Rule) (*Service, error) {
	return &Service{
		autounlocker: autounlocker,
		checker:      checker,
		stores:       stores,
		rules:        rules,
	}, nil
}

// ServeGRPC the wallet service over GRPC.
func (s *Service) ServeGRPC(ctx context.Context, config *core.ServerConfig) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "wallet.service.ServeGRPC")
	defer span.Finish()

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

	var ruler ruler.Service
	if len(s.rules) > 0 {
		log.Info().Int("rules", len(s.rules)).Msg("Enabling rule scripts")
		ruler, err = lua.New(locker, store, s.rules)
	} else {
		log.Info().Msg("Enabling static rules")
		ruler, err = golang.New(locker, store)
	}
	if err != nil {
		return err
	}

	fetcher, err := memfetcher.New(s.stores)
	if err != nil {
		return err
	}

	signerSvc, err := signersvc.New(s.autounlocker, s.checker, fetcher, ruler)
	if err != nil {
		return err
	}

	pb.RegisterWalletManagerServer(s.grpcServer, walletmanager.New(fetcher, ruler))
	pb.RegisterAccountManagerServer(s.grpcServer, accountmanager.New(fetcher, ruler))
	pb.RegisterListerServer(s.grpcServer, lister.New(s.checker, fetcher, ruler))
	pb.RegisterSignerServer(s.grpcServer, signerhandler.New(signerSvc))

	err = s.Serve(config)
	if err != nil {
		return err
	}
	return nil
}

// createServer creates the GRPC server.
func (s *Service) createServer(config *core.ServerConfig) error {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger = logger.With().Str("module", "grpc-wallet").Logger()
	grpclog.SetLoggerV2(util.NewLogShim(logger))

	grpcOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				interceptors.SourceIPInterceptor(),
				interceptors.ClientInfoInterceptor(),
			)),
	}

	if config.Name == "" {
		return errors.New("No server name provided; cannot proceed")
	}

	// Read in the server certificate; this is required to provide security over incoming connections.
	serverCertFile := filepath.Join(config.CertPath, fmt.Sprintf("%s.crt", config.Name))
	log.Debug().Str("path", serverCertFile).Msg("Server certificate file")
	serverKeyFile := filepath.Join(config.CertPath, fmt.Sprintf("%s.key", config.Name))
	log.Debug().Str("path", serverKeyFile).Msg("Server key file")
	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		return errors.Wrap(err, "Failed to load server keypair")
	}

	// Read in the certificate authority certificate; this is required to validate client certificates on incoming connections.
	caCertFile := filepath.Join(config.CertPath, "ca.crt")
	log.Debug().Str("path", caCertFile).Msg("CA certificate file")
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
	log.Info().Str("address", listenAddress).Msg("Listening")

	if err := s.grpcServer.Serve(conn); err != nil {
		return fmt.Errorf("could not serve gRPC: %v", err)
	}
	return nil
}
