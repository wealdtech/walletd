package lister

import (
	context "context"
	"fmt"
	"regexp"
	"strings"

	"github.com/wealdtech/walletd/backend"
	"github.com/wealdtech/walletd/interceptors"
	pb "github.com/wealdtech/walletd/pb/v1"
	"github.com/wealdtech/walletd/util"
	lua "github.com/yuin/gopher-lua"
)

// ListAccounts lists accouts.
func (s *Service) ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
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

		wallet, err := s.fetcher.FetchWallet(path)
		if err != nil {
			log.WithError(err).Info("Failed to obtain wallet")
			continue
		}

		for account := range wallet.Accounts() {
			if accountRegex.Match([]byte(account.Name())) {
				// Confirm listing of the key
				accountName := fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
				rules := s.ruler.Rules("list account", accountName)
				result := backend.RunRules(rules,
					func(l *lua.LState) (*lua.LTable, error) {
						table := l.NewTable()
						table.RawSetString("account", lua.LString(accountName))
						table.RawSetString("pubKey", lua.LString(fmt.Sprintf("%0x", account.PublicKey().Marshal())))
						if ip, ok := ctx.Value(&interceptors.ExternalIP{}).(string); ok {
							table.RawSetString("ip", lua.LString(ip))
						}
						return table, nil
					},
					func() (*backend.State, error) {
						return s.storage.FetchListAccountsState(account.PublicKey().Marshal())
					},
					func(table *lua.LTable, state *backend.State) error {
						table.ForEach(func(k, v lua.LValue) {
							state.Store(k.String(), v)
						})
						return s.storage.StoreListAccountsState(account.PublicKey().Marshal(), state)
					})
				if result == backend.APPROVED {
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
