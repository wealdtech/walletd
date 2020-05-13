package golang_test

import (
	"context"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/wealdtech/eth2-signer-api/pb/v1"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/locker"
	"github.com/wealdtech/walletd/services/ruler"
	"github.com/wealdtech/walletd/services/ruler/golang"
	"github.com/wealdtech/walletd/services/storage/mem"
)

func _byteStr(t *testing.T, input string) []byte {
	bytes, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	require.Nil(t, err)
	return bytes
}

func TestRunRulesSign(t *testing.T) {
	locker, err := locker.New()
	require.NoError(t, err)
	store, err := mem.New()
	require.NoError(t, err)

	r, err := golang.New(locker, store)
	require.NoError(t, err)

	tests := []struct {
		name string
		req  *pb.SignRequest
		res  core.RulesResult
	}{
		{
			name: "SignBeaconAttestationDomain",
			req: &pb.SignRequest{
				Id: &pb.SignRequest_PublicKey{
					PublicKey: _byteStr(t, "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1d1f2021222324252627"),
				},
				Data:   _byteStr(t, "0000000000000000000000000000000000000000000000000000000000000000"),
				Domain: _byteStr(t, "0100000000000000000000000000000000000000000000000000000000000000"),
			},
			res: core.DENIED,
		},
		{
			name: "SignBeaconProposalDomain",
			req: &pb.SignRequest{
				Id: &pb.SignRequest_PublicKey{
					PublicKey: _byteStr(t, "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1d1f2021222324252627"),
				},
				Data:   _byteStr(t, "0000000000000000000000000000000000000000000000000000000000000000"),
				Domain: _byteStr(t, "0000000000000000000000000000000000000000000000000000000000000000"),
			},
			res: core.DENIED,
		},
		{
			name: "Good",
			req: &pb.SignRequest{
				Id: &pb.SignRequest_PublicKey{
					PublicKey: _byteStr(t, "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1d1f2021222324252627"),
				},
				Data:   _byteStr(t, "0000000000000000000000000000000000000000000000000000000000000000"),
				Domain: _byteStr(t, "0200000000000000000000000000000000000000000000000000000000000000"),
			},
			res: core.APPROVED,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := r.RunRules(context.Background(), ruler.ActionSign, "Test wallet", "Test account", test.req.GetPublicKey(), test.req)
			assert.Equal(t, test.res, res)
		})
	}
}
