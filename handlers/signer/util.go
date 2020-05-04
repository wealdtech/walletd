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
