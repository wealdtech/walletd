package lister

import (
	context "context"
	"fmt"
	"regexp"
	"strings"

	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/util"
)

// ListAccounts lists accouts.
func (h *Handler) ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	res := &pb.ListAccountsResponse{}
	res.Accounts = make([]*pb.Account, 0)

	for _, path := range req.Paths {
		log := log.WithField("path", path)
		walletName, accountPath, err := util.WalletAndAccountNamesFromPath(path)
		if err != nil {
			log.WithError(err).Info("Failed to obtain wallet and accout names")
			continue
		}
		if walletName == "" {
			log.Info("Empty wallet name")
			continue
		}

		if accountPath == "" {
			accountPath = "^.*$"
		}
		if !strings.HasPrefix(accountPath, "^") {
			accountPath = fmt.Sprintf("^%s", accountPath)
		}
		if !strings.HasSuffix(accountPath, "$") {
			accountPath = fmt.Sprintf("%s$", accountPath)
		}
		accountRegex, err := regexp.Compile(accountPath)
		if err != nil {
			log.WithError(err).Info("Invalid account regular expression")
		}

		wallet, err := h.fetcher.FetchWallet(path)
		if err != nil {
			log.WithError(err).Info("Failed to obtain wallet")
			continue
		}

		for account := range wallet.Accounts() {
			if accountRegex.Match([]byte(account.Name())) {
				// Confirm listing of the key.
				result := h.ruler.RunRules(ctx, "list account", wallet, account, nil)
				if result == core.APPROVED {
					res.Accounts = append(res.Accounts, &pb.Account{
						Name:      fmt.Sprintf("%s/%s", wallet.Name(), account.Name()),
						PublicKey: account.PublicKey().Marshal(),
					})
				}
			}
		}
	}
	return res, nil
}
