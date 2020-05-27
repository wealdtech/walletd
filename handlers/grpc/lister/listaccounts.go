// Copyright © 2020 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lister

import (
	context "context"
	"fmt"
	"regexp"
	"strings"

	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/interceptors"
	"github.com/wealdtech/walletd/services/checker"
	"github.com/wealdtech/walletd/services/ruler"
	"github.com/wealdtech/walletd/util"
)

// ListAccounts lists accounts.
func (h *Handler) ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	log.Info().Strs("paths", req.GetPaths()).Msg("List accounts request received")
	res := &pb.ListAccountsResponse{}
	res.Accounts = make([]*pb.Account, 0)

	for _, path := range req.Paths {
		log := log.With().Str("path", path).Logger()
		walletName, accountPath, err := util.WalletAndAccountNamesFromPath(path)
		if err != nil {
			log.Info().Err(err).Msg("Failed to obtain wallet and account names from path")
			continue
		}
		if walletName == "" {
			log.Info().Msg("Empty wallet in path")
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
			log.Info().Err(err).Msg("Invalid account regular expression")
			continue
		}

		wallet, err := h.fetcher.FetchWallet(path)
		if err != nil {
			log.Info().Err(err).Msg("Failed to obtain wallet")
			continue
		}

		for account := range wallet.Accounts() {
			if accountRegex.Match([]byte(account.Name())) {
				accountName := fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
				// Confirm access to the key.
				ok, err := h.checkClientAccess(ctx, accountName, ruler.ActionAccessAccount)
				if err != nil {
					log.Warn().Err(err).Msg("Failed to check account for access")
					continue
				}
				if !ok {
					// Not allowed.
					continue
				}

				// Confirm listing of the key.
				pubKey := account.PublicKey().Marshal()
				result := h.ruler.RunRules(ctx, ruler.ActionAccessAccount, wallet.Name(), account.Name(), pubKey, req)
				if result == core.APPROVED {
					res.Accounts = append(res.Accounts, &pb.Account{
						Name:      accountName,
						PublicKey: pubKey,
					})
				}
			}
		}
	}

	res.State = pb.ResponseState_SUCCEEDED
	log.Info().Msg("Success")
	return res, nil
}

// checkClientAccess returns true if the client can access the account.
func (h *Handler) checkClientAccess(ctx context.Context, accountName string, operation string) (bool, error) {
	client := ""
	val := ctx.Value(&interceptors.ClientName{})
	if val != nil {
		client = val.(string)
	}
	return h.checker.Check(ctx, &checker.Credentials{Client: client}, accountName, operation), nil
}
