package wallet

import (
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/backend"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/handlers/accountmanager"
	"github.com/wealdtech/walletd/handlers/lister"
	"github.com/wealdtech/walletd/handlers/signer"
	"github.com/wealdtech/walletd/handlers/walletmanager"
	"github.com/wealdtech/walletd/services/ruler/lua"
	"github.com/wealdtech/walletd/services/storage/badger"
)

// Service provides the features and functions for the wallet daemon.
type Service struct {
	stores []e2wtypes.Store
	rules  []*core.Rule
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
	grpcServer, err := core.NewServer(config)
	if err != nil {
		return err
	}

	store, err := badger.New(config.StoragePath)
	if err != nil {
		return err
	}

	ruler, err := lua.New(store, s.rules)
	if err != nil {
		return err
	}

	fetcher := backend.NewMemFetcher(s.stores)

	pb.RegisterWalletManagerServer(grpcServer.Server(), walletmanager.New(fetcher, ruler))
	pb.RegisterAccountManagerServer(grpcServer.Server(), accountmanager.New(fetcher, ruler))
	pb.RegisterListerServer(grpcServer.Server(), lister.New(fetcher, ruler, store))
	pb.RegisterSignerServer(grpcServer.Server(), signer.New(fetcher, ruler, store))

	err = grpcServer.Serve()
	if err != nil {
		return err
	}
	return nil
}
