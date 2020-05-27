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
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/checker"
	"github.com/wealdtech/walletd/services/ruler"
)

// Sign signs generic data.
func (s *Service) Sign(ctx context.Context, credentials *checker.Credentials, accountName string, pubKey []byte, data *ruler.SignData) (core.RulesResult, []byte) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.signer.Sign")
	defer span.Finish()
	log := log.With().Str("action", "Sign").Logger()
	log.Debug().Msg("Request received")

	if data == nil {
		return core.DENIED, nil
	}
	wallet, account, checkRes := s.preCheck(ctx, credentials, accountName, pubKey, ruler.ActionSign)
	if checkRes != core.APPROVED {
		return checkRes, nil
	}
	accountName = fmt.Sprintf("%s/%s", wallet.Name(), account.Name())
	log = log.With().Str("account", accountName).Logger()

	// Confirm approval via rules.
	result := s.ruler.RunRules(ctx, ruler.ActionSign, wallet.Name(), account.Name(), account.PublicKey().Marshal(), data)
	switch result {
	case core.DENIED:
		log.Debug().Str("result", "denied").Msg("Denied by rules")
		return core.DENIED, nil
	case core.FAILED:
		log.Warn().Str("result", "failed").Msg("Rules check failed")
		return core.FAILED, nil
	}

	// Sign it.
	signingRoot, err := generateSigningRootFromRoot(ctx, data.Data, data.Domain)
	if err != nil {
		log.Warn().Err(err).Str("result", "failed").Msg("Failed to generate signing root")
		return core.FAILED, nil
	}
	span, _ = opentracing.StartSpanFromContext(ctx, "service.signer.Sign/Sign")
	signature, err := account.Sign(signingRoot[:])
	if err != nil {
		log.Warn().Err(err).Str("result", "failed").Msg("Failed to sign")
		span.Finish()
		return core.FAILED, nil
	}
	span.Finish()

	log.Debug().Str("result", "succeeded").Msg("Success")
	return core.APPROVED, signature.Marshal()
}
