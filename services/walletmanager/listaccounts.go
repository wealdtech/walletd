package walletmanager

import (
	context "context"
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	pb "github.com/wealdtech/walletd/pb/v1"
	"github.com/wealdtech/walletd/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListAccounts lists accounts given a path.
func (s *Service) ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	log.WithField("path", req.Path).Debug("ListAccounts()")
	walletName, accountName, err := util.WalletAndAccountNamesFromPath(req.Path)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if walletName == "" {
		return nil, status.Error(codes.InvalidArgument, "No wallet name supplied")
	}

	wallet, err := s.fetcher.FetchWallet(req.Path)
	if err != nil {
		return nil, status.Error(codes.NotFound, "No such wallet")
	}

	accRegex, err := regexp.Compile(fmt.Sprintf("^%s$", accountName))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid regular expression")
	}

	accounts := make([]*pb.Account, 0)
	for account := range wallet.Accounts() {
		if accRegex.Match([]byte(account.Name())) {
			uuid, err := account.ID().MarshalBinary()
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"wallet":  walletName,
					"account": account.Name(),
				}).Error("Failed to marshal UUID; skipping")
				continue
			}

			accounts = append(accounts, &pb.Account{
				Uuid:      uuid,
				Name:      account.Name(),
				PublicKey: account.PublicKey().Marshal(),
			})
		}
	}

	return &pb.ListAccountsResponse{Accounts: accounts}, nil
}
