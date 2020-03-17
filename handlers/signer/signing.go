package signer

import (
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// fetchAccount is a utility that fetches an account from either account name or public key.
func (h *Handler) fetchAccount(name string, pubKey []byte) (e2wtypes.Wallet, e2wtypes.Account, error) {
	if name == "" {
		return h.fetcher.FetchAccountByKey(pubKey)
	}
	return h.fetcher.FetchAccount(name)
}
