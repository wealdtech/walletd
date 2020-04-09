package signer

import (
	ssz "github.com/prysmaticlabs/go-ssz"
)

func generateSigningRootFromData(data interface{}, domain []byte) ([32]byte, error) {
	objRoot, err := ssz.HashTreeRoot(data)
	if err != nil {
		return [32]byte{}, err
	}

	return generateSigningRootFromRoot(objRoot[:], domain)
}

func generateSigningRootFromRoot(root []byte, domain []byte) ([32]byte, error) {
	signingData := struct {
		Hash   []byte `ssz-size:"32"`
		Domain []byte `ssz-size:"32"`
	}{
		Hash:   root,
		Domain: domain,
	}
	return ssz.HashTreeRoot(signingData)
}
