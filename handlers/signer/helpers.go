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
	"errors"

	"github.com/opentracing/opentracing-go"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/interceptors"
)

// fetchAccount is a utility that fetches an account from either account name or public key.
func (h *Handler) fetchAccount(ctx context.Context, name string, pubKey []byte) (e2wtypes.Wallet, e2wtypes.Account, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "handlers.signer.fetchAccount")
	defer span.Finish()

	if name == "" {
		return h.fetcher.FetchAccountByKey(pubKey)
	}
	return h.fetcher.FetchAccount(name)
}

// checkClientAccess returns true if the client can access the account.
func (h *Handler) checkClientAccess(ctx context.Context, accountName string, operation string) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "handlers.signer.checkClientAccess")
	defer span.Finish()

	client, ok := ctx.Value(&interceptors.ClientName{}).(string)
	if !ok {
		return false, errors.New("no client certificate name")
	}
	return h.checker.Check(ctx, string(client), accountName, operation), nil
}
