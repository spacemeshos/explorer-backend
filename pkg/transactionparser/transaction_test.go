package transactionparser_test

import (
	"bytes"
	"math/rand"
	"testing"
	"time"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/spacemeshos/go-scale"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	"github.com/spacemeshos/go-spacemesh/genvm/sdk"
	sdkMultisig "github.com/spacemeshos/go-spacemesh/genvm/sdk/multisig"
	sdkWallet "github.com/spacemeshos/go-spacemesh/genvm/sdk/wallet"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/multisig"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"
	"github.com/spacemeshos/go-spacemesh/signing"
	"github.com/stretchr/testify/require"

	"github.com/spacemeshos/explorer-backend/pkg/transactionparser"
)

func TestSpawn(t *testing.T) {
	table := []struct {
		name     string
		gasPrice uint64
		opts     []sdk.Opt
	}{
		{
			name:     "default gas price",
			gasPrice: 1,
			opts:     []sdk.Opt{},
		},
		{
			name:     "non default gasPrice",
			gasPrice: 2,
			opts:     []sdk.Opt{sdk.WithGasPrice(2)},
		},
	}
	for _, tc := range table {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			signer := signing.NewEdSigner()
			rawTx := sdkWallet.SelfSpawn(signer.PrivateKey(), testCase.opts...)

			args := wallet.SpawnArguments{}
			copy(args.PublicKey[:], signer.PublicKey().Bytes())
			principal := core.ComputePrincipal(wallet.TemplateAddress, &args)

			decodedTx, err := transactionparser.Parse(scale.NewDecoder(bytes.NewReader(rawTx)), rawTx, 0)
			require.NoError(t, err)
			require.Equal(t, testCase.gasPrice, decodedTx.GetGasPrice())
			require.Equal(t, principal.String(), decodedTx.GetPrincipal().String())
		})
	}
}

func TestSpend(t *testing.T) {
	table := []struct {
		name     string
		to       types.Address
		gasPrice uint64
		amount   uint64
		opts     []sdk.Opt
		nonce    types.Nonce
	}{
		{
			name:     "default gas price",
			amount:   123,
			gasPrice: 1,
			to:       types.GenerateAddress(generatePublicKey(t)),
			opts:     []sdk.Opt{},
			nonce: types.Nonce{
				Counter:  123,
				Bitfield: uint8(1),
			},
		},
		{
			name:     "non default gasPrice",
			amount:   723,
			gasPrice: 2,
			to:       types.GenerateAddress(generatePublicKey(t)),
			opts:     []sdk.Opt{sdk.WithGasPrice(2)},
			nonce: types.Nonce{
				Counter:  723,
				Bitfield: uint8(4),
			},
		},
	}
	for _, tc := range table {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			signer := signing.NewEdSigner()
			rawTx := sdkWallet.Spend(signer.PrivateKey(), testCase.to, testCase.amount, testCase.nonce, testCase.opts...)

			args := wallet.SpawnArguments{}
			copy(args.PublicKey[:], signing.Public(signer.PrivateKey()))
			accAddress := core.ComputePrincipal(wallet.TemplateAddress, &args)

			decodedTx, err := transactionparser.Parse(scale.NewDecoder(bytes.NewReader(rawTx)), rawTx, 1)
			require.NoError(t, err)
			require.Equal(t, testCase.gasPrice, decodedTx.GetGasPrice())
			require.Equal(t, testCase.to.String(), decodedTx.GetReceiver().String())
			require.Equal(t, testCase.amount, decodedTx.GetAmount())
			require.Equal(t, testCase.nonce.Counter, decodedTx.GetCounter())
			require.Equal(t, accAddress.String(), decodedTx.GetPrincipal().String())
		})
	}
}

func TestSpawnMultisig(t *testing.T) {
	table := []struct {
		name     string
		gasPrice uint64
		ref      uint8
		opts     []sdk.Opt
	}{
		{
			name:     "default gas price",
			gasPrice: 1,
			ref:      3,
			opts:     []sdk.Opt{},
		},
		{
			name:     "non default gasPrice",
			gasPrice: 2,
			ref:      5,
			opts:     []sdk.Opt{sdk.WithGasPrice(2)},
		},
	}
	for _, tc := range table {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var pubs []ed25519.PublicKey
			pks := make([]ed25519.PrivateKey, 0, 3)
			for i := 0; i < 3; i++ {
				pub, pk, err := ed25519.GenerateKey(rand.New(rand.NewSource(time.Now().UnixNano())))
				require.NoError(t, err)
				pubs = append(pubs, pub)
				pks = append(pks, pk)
			}

			var agg *sdkMultisig.Aggregator
			for i := 0; i < len(pks); i++ {
				part := sdkMultisig.SelfSpawn(uint8(i), pks[i], multisig.TemplateAddress3, pubs, testCase.opts...)
				if agg == nil {
					agg = part
				} else {
					agg.Add(*part.Part(uint8(i)))
				}
			}
			rawTx := agg.Raw()

			decodedTx, err := transactionparser.Parse(scale.NewDecoder(bytes.NewReader(rawTx)), rawTx, 0)
			require.NoError(t, err)
			require.Equal(t, testCase.gasPrice, decodedTx.GetGasPrice())
			for i := 0; i < len(pks); i++ {
				//	require.Equal(t, []byte(pubs[i]), decodedTx.MultisigSpawnArgs.PublicKeys[i].Bytes())
			}
		})
	}
}

func TestSpendMultisig(t *testing.T) {
	table := []struct {
		name     string
		to       types.Address
		gasPrice uint64
		amount   uint64
		opts     []sdk.Opt
		nonce    types.Nonce
		ref      uint8
	}{
		{
			name:     "default gas price",
			amount:   123,
			gasPrice: 1,
			to:       types.GenerateAddress(generatePublicKey(t)),
			opts:     []sdk.Opt{},
			ref:      3,
			nonce: types.Nonce{
				Counter:  123,
				Bitfield: uint8(1),
			},
		},
		{
			name:     "non default gasPrice",
			amount:   723,
			gasPrice: 2,
			ref:      5,
			to:       types.GenerateAddress(generatePublicKey(t)),
			opts:     []sdk.Opt{sdk.WithGasPrice(2)},
			nonce: types.Nonce{
				Counter:  723,
				Bitfield: uint8(4),
			},
		},
	}
	for _, tc := range table {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pubs := make([][]byte, 0, 3)
			pks := make([]ed25519.PrivateKey, 0, 3)
			for i := 0; i < 3; i++ {
				pub, pk, err := ed25519.GenerateKey(rand.New(rand.NewSource(time.Now().UnixNano())))
				require.NoError(t, err)
				pubs = append(pubs, pub)
				pks = append(pks, pk)
			}
			principal := sdkMultisig.Address(multisig.TemplateAddress3, pubs...)
			agg := sdkMultisig.Spend(0, pks[0], principal, testCase.to, testCase.amount, testCase.nonce, testCase.opts...)
			for i := 1; i < len(pks); i++ {
				part := sdkMultisig.Spend(uint8(i), pks[i], principal, testCase.to, testCase.amount, testCase.nonce, testCase.opts...)
				agg.Add(*part.Part(uint8(i)))
			}
			rawTx := agg.Raw()

			decodedTx, err := transactionparser.Parse(scale.NewDecoder(bytes.NewReader(rawTx)), rawTx, 1)
			require.NoError(t, err)
			require.Equal(t, testCase.gasPrice, decodedTx.GetGasPrice())
			require.Equal(t, testCase.to.String(), decodedTx.GetReceiver().String())
			require.Equal(t, testCase.amount, decodedTx.GetAmount())
			require.Equal(t, testCase.nonce.Counter, decodedTx.GetCounter())
			require.Equal(t, principal.String(), decodedTx.GetPrincipal().String())
		})
	}
}

func generatePublicKey(t *testing.T) []byte {
	buff := signing.NewEdSigner().ToBuffer()
	acc1Signer, err := signing.NewEdSignerFromBuffer(buff)
	require.NoError(t, err)
	return acc1Signer.PublicKey().Bytes()
}
