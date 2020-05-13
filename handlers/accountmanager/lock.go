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

package accountmanager

import (
	context "context"

	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
)

// Lock locks an account.
func (h *Handler) Lock(ctx context.Context, req *pb.LockAccountRequest) (*pb.LockAccountResponse, error) {
	log.Info().Str("account", req.GetAccount()).Msg("Lock account received")
	res := &pb.LockAccountResponse{}

	_, account, err := h.fetcher.FetchAccount(req.Account)
	if err != nil {
		log.Info().Err(err).Str("result", "denied").Msg("Failed to fetch account")
		res.State = pb.ResponseState_DENIED
	} else {
		account.Lock()
		log.Info().Str("result", "succeeded").Msg("Account locked")
		res.State = pb.ResponseState_SUCCEEDED
	}
	return res, nil
}
