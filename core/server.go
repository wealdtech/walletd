package core

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/wealdtech/walletd/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server is a server for GRPC.
type Server struct {
	grpcServer *grpc.Server
	config     *ServerConfig
}

// NewServer creates a new GRPC server.
func NewServer(config *ServerConfig) (*Server, error) {
	grpcOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				interceptors.SourceIPInterceptor(),
				grpc_logrus.UnaryServerInterceptor(log.NewEntry(log.StandardLogger())),
				interceptors.ClientInfoInterceptor(),
			)),
	}

	// Read in the server certificate; this is required to provide security over incoming connections.
	serverCertFile := filepath.Join(config.CertPath, fmt.Sprintf("%s.crt", config.Name))
	serverKeyFile := filepath.Join(config.CertPath, fmt.Sprintf("%s.key", config.Name))
	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to load server keypair")
	}

	// Read in the certificate authority certificate; this is required to validate client certificates on incoming connections.
	caCertFile := filepath.Join(config.CertPath, "ca.crt")
	certPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Could not read CA certificate from %s", caCertFile))
	}
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return nil, errors.Wrap(err, "Could not add CA certificate to pool")
	}

	serverCreds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    certPool,
	})
	grpcOpts = append(grpcOpts, grpc.Creds(serverCreds))
	grpcServer := grpc.NewServer(grpcOpts...)

	return &Server{
		grpcServer: grpcServer,
		config:     config,
	}, nil
}

// Server returns the underlying GRPC server.
func (s *Server) Server() *grpc.Server {
	return s.grpcServer
}

// RegisterService registers a GRPC service.
func (s *Server) RegisterService() error {
	return nil
}

// Serve serves the GRPC server.
func (s *Server) Serve() error {
	listenAddress := fmt.Sprintf(":%d", s.config.Port)
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
