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

	"github.com/opentracing/opentracing-go"
	ssz "github.com/prysmaticlabs/go-ssz"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func generateSigningRootFromData(ctx context.Context, data interface{}, domain []byte) ([32]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "handlers.signer.generateSigningRootFromData")
	defer span.Finish()

	objRoot, err := ssz.HashTreeRoot(data)
	if err != nil {
		return [32]byte{}, err
	}

	return generateSigningRootFromRoot(ctx, objRoot[:], domain)
}

func generateSigningRootFromRoot(ctx context.Context, root []byte, domain []byte) ([32]byte, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "handlers.signer.generateSigningRootFromRoot")
	defer span.Finish()

	signingData := struct {
		Hash   []byte `ssz-size:"32"`
		Domain []byte `ssz-size:"32"`
	}{
		Hash:   root,
		Domain: domain,
	}
	return ssz.HashTreeRoot(signingData)
}

func signRoot(ctx context.Context, account e2wtypes.Account, root []byte) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "handlers.signer.signRoot")
	defer span.Finish()

	span, _ = opentracing.StartSpanFromContext(ctx, "handlers.signer.signRoot/Sign")
	signature, err := account.Sign(root)
	span.Finish()
	if err != nil {
		return nil, err
	}
	span, _ = opentracing.StartSpanFromContext(ctx, "handlers.signer.signRoot/Marshal")
	defer span.Finish()
	return signature.Marshal(), nil
}
