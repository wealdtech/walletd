package core

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

// Server is a server for GRPC.
type Server struct {
	grpcServer *grpc.Server
}

// NewServer creates a new GRPC server.
func NewServer() (*Server, error) {
	grpcServer := grpc.NewServer()

	return &Server{
		grpcServer: grpcServer,
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
	conn, err := net.Listen("tcp", ":12346")
	if err != nil {
		return err
	}

	if err := s.grpcServer.Serve(conn); err != nil {
		return fmt.Errorf("could not serve gRPC: %v", err)
	}
	return nil
}
