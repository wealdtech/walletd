package lister

import (
	context "context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/interceptors"
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
				// Confirm access to the key
				ok, err := h.checkClientAccess(ctx, wallet, account, "Access account")
				if err != nil {
					log.WithError(err).Warn("Failed to check account")
					continue
				}
				if !ok {
					// Not allowed
					continue
				}

				// Confirm listing of the key.
				result := h.ruler.RunRules(ctx, "Access account", wallet, account, nil)
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

// checkClientAccess returns true if the client can access the account.
func (h *Handler) checkClientAccess(ctx context.Context, wallet e2wtypes.Wallet, account e2wtypes.Account, operation string) (bool, error) {
	client, ok := ctx.Value(&interceptors.ClientName{}).(string)
	if !ok {
		return false, errors.New("no client certificate name")
	}
	accountName := fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
	return h.checker.Check(string(client), accountName, operation), nil
}
