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

package signer

import (
	context "context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/checker"
)

// preCheck carries out pre-checks for all signing requests.
func (s *Service) preCheck(ctx context.Context, credentials *checker.Credentials, name string, pubKey []byte, action string) (e2wtypes.Wallet, e2wtypes.Account, core.RulesResult) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "services.signer.preCheck")
	defer span.Finish()

	// Fetch the account.
	wallet, account, result := s.fetchAccount(ctx, credentials, name, pubKey)
	if result != core.APPROVED {
		return nil, nil, result
	}
	accountName := fmt.Sprintf("%s/%s", wallet.Name(), account.Name())

	// Check if the account is allowed to carry out the requested action.
	result = s.checkAccess(ctx, credentials, accountName, action)
	if result != core.APPROVED {
		return nil, nil, result
	}

	// Unlock the account if necessary.
	result = s.unlockAccount(ctx, wallet, account)
	if result != core.APPROVED {
		return nil, nil, result
	}

	return wallet, account, core.APPROVED
}

// fetchAccount fetches an account by either name or public key, depending on which has been supplied.
func (s *Service) fetchAccount(ctx context.Context, credentials *checker.Credentials, name string, pubKey []byte) (e2wtypes.Wallet, e2wtypes.Account, core.RulesResult) {
	span, _ := opentracing.StartSpanFromContext(ctx, "services.signer.fetchAccount")
	defer span.Finish()

	if name == "" && pubKey == nil {
		log.Debug().Str("result", "denied").Msg("Neither account nor public key supplied; denied")
		return nil, nil, core.DENIED
	}

	var wallet e2wtypes.Wallet
	var account e2wtypes.Account
	var err error
	if pubKey == nil {
		wallet, account, err = s.fetcher.FetchAccount(name)
	} else {
		wallet, account, err = s.fetcher.FetchAccountByKey(pubKey)
	}

	if err != nil {
		log.Debug().Err(err).Str("result", "denied").Msg("Did not obtain account; denied")
		return nil, nil, core.DENIED
	}

	return wallet, account, core.APPROVED
}

// checkAccess returns true if the client can access the account.
func (s *Service) checkAccess(ctx context.Context, credentials *checker.Credentials, accountName string, action string) core.RulesResult {
	span, ctx := opentracing.StartSpanFromContext(ctx, "services.signer.checkAccess")
	defer span.Finish()

	if s.checker.Check(ctx, credentials, accountName, action) {
		return core.APPROVED
	}
	return core.DENIED
}

// unlockAccount returns true if the client can access the account.
func (s *Service) unlockAccount(ctx context.Context, wallet e2wtypes.Wallet, account e2wtypes.Account) core.RulesResult {
	span, ctx := opentracing.StartSpanFromContext(ctx, "services.signer.accountUnlock")
	defer span.Finish()

	if wallet == nil {
		log.Debug().Str("result", "denied").Msg("No wallet provided")
		return core.DENIED
	}
	if account == nil {
		log.Debug().Str("result", "denied").Msg("No account provided")
		return core.DENIED
	}

	if account.IsUnlocked() {
		return core.APPROVED
	}

	unlocked, err := s.autounlocker.Unlock(ctx, wallet, account)
	if err != nil {
		log.Debug().Str("result", "failed").Msg("Failed during attempt to unlock account")
		return core.FAILED
	}
	if !unlocked {
		log.Debug().Str("result", "denied").Msg("Account is locked; signing request denied")
		return core.DENIED
	}
	return core.APPROVED
}
