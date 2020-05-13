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

package keys

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/walletd/core"
)

// Service is an autounlocker service that holds unlock passphrases.
type Service struct {
	passphrases [][]byte
}

// New creates a new autounlocker service that holds unlock passphrases.
func New(ctx context.Context, config *core.KeysConfig) (*Service, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "autounlocker.keys.New")
	defer span.Finish()

	passphrases := make([][]byte, len(config.Keys))
	for i, key := range config.Keys {
		passphrases[i] = []byte(key)
	}

	return &Service{
		passphrases: passphrases,
	}, nil
}

// Unlock attempts to unlock an account.
func (s *Service) Unlock(ctx context.Context, wallet e2wtypes.Wallet, account e2wtypes.Account) (bool, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "autounlocker.keys.Unlock")
	defer span.Finish()

	if account == nil {
		return false, errors.New("no account supplied")
	}

	for _, passphrase := range s.passphrases {
		if err := account.Unlock(passphrase); err == nil {
			return true, nil
		}
	}
	return false, nil
}
