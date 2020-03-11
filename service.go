package main

import (
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/backend"
	"github.com/wealdtech/walletd/core"
	pb "github.com/wealdtech/walletd/pb/v1"
	"github.com/wealdtech/walletd/services/accountmanager"
	"github.com/wealdtech/walletd/services/lister"
	"github.com/wealdtech/walletd/services/signer"
	"github.com/wealdtech/walletd/services/walletmanager"
)

// WalletService provides the features and functions for the wallet.
type WalletService struct {
	stores []e2wtypes.Store
	rules  []*core.Rule
}

// NewWalletService creates a new wallet.
func NewWalletService(stores []e2wtypes.Store, rules []*core.Rule) (*WalletService, error) {
	return &WalletService{
		stores: stores,
		rules:  rules,
	}, nil
}

// ServeGRPC the wallet service over GRPC.
func (w *WalletService) ServeGRPC() error {
	grpcServer, err := core.NewServer()
	if err != nil {
		return err
	}

	fetcher := backend.NewMemFetcher(w.stores)
	ruler := backend.NewStaticRuler(w.rules)
	storage, err := backend.NewFSStorage("TODO")
	if err != nil {
		return err
	}

	pb.RegisterWalletManagerServer(grpcServer.Server(), walletmanager.NewService(fetcher, ruler))
	pb.RegisterAccountManagerServer(grpcServer.Server(), accountmanager.NewService(fetcher, ruler))
	pb.RegisterListerServer(grpcServer.Server(), lister.NewService(fetcher, ruler, storage))
	pb.RegisterSignerServer(grpcServer.Server(), signer.NewService(fetcher, ruler, storage))

	err = grpcServer.Serve()
	if err != nil {
		return err
	}
	return nil
}
