package signer

import (
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// fetchAccount fetches an account.
func (s *Service) fetchAccount(name string, pubKey []byte) (e2wtypes.Wallet, e2wtypes.Account, error) {
	if name == "" {
		return s.fetcher.FetchAccountByKey(pubKey)
	}
	return s.fetcher.FetchAccount(name)
}
