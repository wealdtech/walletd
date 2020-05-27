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

package walletmanager

import (
	context "context"

	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
)

// Lock locks a wallet.
func (h *Handler) Lock(ctx context.Context, req *pb.LockWalletRequest) (*pb.LockWalletResponse, error) {
	log.Info().Str("wallet", req.GetWallet()).Msg("Lock wallet received")
	res := &pb.LockWalletResponse{}

	wallet, err := h.fetcher.FetchWallet(req.Wallet)
	if err != nil {
		log.Info().Err(err).Str("result", "denied").Msg("Failed to fetch wallet")
		res.State = pb.ResponseState_DENIED
	} else {
		wallet.Lock()
		log.Info().Str("result", "succeeded").Msg("Wallet locked")
		res.State = pb.ResponseState_SUCCEEDED
	}
	return res, nil
}
