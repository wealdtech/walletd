package main

import (
	wtypes "github.com/wealdtech/go-eth2-wallet-types"
	"github.com/wealdtech/walletd/backend"
	"github.com/wealdtech/walletd/core"
	pb "github.com/wealdtech/walletd/pb/v1"
	"github.com/wealdtech/walletd/services/accountmanager"
	"github.com/wealdtech/walletd/services/signer"
	"github.com/wealdtech/walletd/services/walletmanager"
)

// WalletService provides the features and functions for the wallet.
type WalletService struct {
	stores []wtypes.Store
	rules  []*core.Rule
}

// NewWalletService creates a new wallet.
func NewWalletService(stores []wtypes.Store, rules []*core.Rule) (*WalletService, error) {
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

	pb.RegisterWalletManagerServer(grpcServer.Server(), walletmanager.NewService(fetcher, ruler))
	pb.RegisterAccountManagerServer(grpcServer.Server(), accountmanager.NewService(fetcher, ruler))
	pb.RegisterSignerServer(grpcServer.Server(), signer.NewService(fetcher, ruler))

	err = grpcServer.Serve()
	if err != nil {
		return err
	}
	return nil
}
