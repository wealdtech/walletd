package main

import (
	wtypes "github.com/wealdtech/go-eth2-wallet-types"
	"github.com/wealdtech/go-grpcserver"
	"github.com/wealdtech/walletd/backend"
	pb "github.com/wealdtech/walletd/pb/v1"
	accountsvc "github.com/wealdtech/walletd/services/account"
	signsvc "github.com/wealdtech/walletd/services/sign"
	walletsvc "github.com/wealdtech/walletd/services/wallet"
)

// WalletService provides the features and functions for the wallet.
type WalletService struct {
	stores []wtypes.Store
}

// NewWalletService creates a new wallet.
func NewWalletService(stores []wtypes.Store) (*WalletService, error) {
	return &WalletService{
		stores: stores,
	}, nil
}

// ServeGRPC the wallet service over GRPC.
func (w *WalletService) ServeGRPC() error {
	grpcServer, err := grpcserver.NewServer()
	if err != nil {
		return err
	}

	fetcher := backend.NewMemFetcher(w.stores)

	pb.RegisterWalletServer(grpcServer.Server(), walletsvc.NewService(fetcher))
	pb.RegisterAccountServer(grpcServer.Server(), accountsvc.NewService(fetcher))
	pb.RegisterSignServer(grpcServer.Server(), signsvc.NewService(fetcher))

	err = grpcServer.Serve()
	if err != nil {
		return err
	}
	return nil
}
