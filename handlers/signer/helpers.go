package signer

import (
	context "context"
	"errors"
	"fmt"

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
func (h *Handler) checkClientAccess(ctx context.Context, wallet e2wtypes.Wallet, account e2wtypes.Account, operation string) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "handlers.signer.checkClientAccess")
	defer span.Finish()

	client, ok := ctx.Value(&interceptors.ClientName{}).(string)
	if !ok {
		return false, errors.New("no client certificate name")
	}
	accountName := fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
	return h.checker.Check(ctx, string(client), accountName, operation), nil
}
