package transactionparser_test

import (
	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/spacemeshos/explorer-backend/pkg/transactionparser/transaction"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	"github.com/spacemeshos/go-spacemesh/genvm/sdk"
	multisig2 "github.com/spacemeshos/go-spacemesh/genvm/sdk/multisig"
	"github.com/spacemeshos/go-spacemesh/genvm/sdk/vesting"
	sdkWallet "github.com/spacemeshos/go-spacemesh/genvm/sdk/wallet"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/multisig"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/vault"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"
	"github.com/spacemeshos/go-spacemesh/signing"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"

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
			signer, _ := signing.NewEdSigner()
			rawTx := sdkWallet.SelfSpawn(signer.PrivateKey(), core.Nonce(0), testCase.opts...)
			args := wallet.SpawnArguments{}
			copy(args.PublicKey[:], signer.PublicKey().Bytes())
			principal := core.ComputePrincipal(wallet.TemplateAddress, &args)

			decodedTx, err := transactionparser.Parse(rawTx)
			require.NoError(t, err)

			require.Equal(t, testCase.gasPrice, decodedTx.Tx.GetGasPrice())
			require.Equal(t, principal.String(), decodedTx.Tx.GetPrincipal().String())
			require.Equal(t, transaction.TypeSpawn, decodedTx.Type)
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
			to:       types.GenerateAddress(generatePublicKey()),
			opts:     []sdk.Opt{},
			nonce:    types.Nonce(0),
		},
		{
			name:     "non default gasPrice",
			amount:   723,
			gasPrice: 2,
			to:       types.GenerateAddress(generatePublicKey()),
			opts:     []sdk.Opt{sdk.WithGasPrice(2)},
			nonce:    types.Nonce(0),
		},
	}
	for _, tc := range table {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			signer, _ := signing.NewEdSigner()
			rawTx := sdkWallet.Spend(signer.PrivateKey(), testCase.to, testCase.amount, testCase.nonce, testCase.opts...)
			args := wallet.SpawnArguments{}
			copy(args.PublicKey[:], signing.Public(signer.PrivateKey()))
			accAddress := core.ComputePrincipal(wallet.TemplateAddress, &args)

			decodedTx, err := transactionparser.Parse(rawTx)
			require.NoError(t, err)

			require.Equal(t, testCase.gasPrice, decodedTx.Tx.GetGasPrice())
			require.Equal(t, testCase.to.String(), decodedTx.Tx.GetReceiver().String())
			require.Equal(t, testCase.amount, decodedTx.Tx.GetAmount())
			require.Equal(t, testCase.nonce, decodedTx.Tx.GetCounter())
			require.Equal(t, accAddress.String(), decodedTx.Tx.GetPrincipal().String())
			require.Equal(t, transaction.TypeSpend, decodedTx.Type)
		})
	}
}

func TestSpawnMultisig(t *testing.T) {
	table := []struct {
		name     string
		gasPrice uint64
		ref      uint8
		opts     []sdk.Opt
		nonce    types.Nonce
	}{
		{
			name:     "default gas price",
			gasPrice: 1,
			ref:      3,
			opts:     []sdk.Opt{},
			nonce:    types.Nonce(1),
		},
		{
			name:     "non default gasPrice",
			gasPrice: 2,
			ref:      5,
			opts:     []sdk.Opt{sdk.WithGasPrice(2)},
			nonce:    types.Nonce(2),
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

			var agg *multisig2.Aggregator
			for i := 0; i < len(pks); i++ {
				part := multisig2.SelfSpawn(uint8(i), pks[i], multisig.TemplateAddress, 1, pubs, testCase.nonce, testCase.opts...)
				if agg == nil {
					agg = part
				} else {
					agg.Add(*part.Part(uint8(i)))
				}
			}
			rawTx := agg.Raw()

			decodedTx, err := transactionparser.Parse(rawTx)
			require.NoError(t, err)
			require.Equal(t, testCase.gasPrice, decodedTx.Tx.GetGasPrice())
			require.Equal(t, transaction.TypeMultisigSpawn, decodedTx.Type)
			for i := 0; i < len(pks); i++ {
				require.Equal(t, []byte(pubs[i]), decodedTx.Tx.GetPublicKeys()[i])
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
			to:       types.GenerateAddress(generatePublicKey()),
			opts:     []sdk.Opt{},
			ref:      3,
			nonce:    types.Nonce(1),
		},
		{
			name:     "non default gasPrice",
			amount:   723,
			gasPrice: 2,
			ref:      5,
			to:       types.GenerateAddress(generatePublicKey()),
			opts:     []sdk.Opt{sdk.WithGasPrice(2)},
			nonce:    types.Nonce(2),
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
			principal := multisig2.Address(multisig.TemplateAddress, 3, pubs...)
			agg := multisig2.Spend(0, pks[0], principal, testCase.to, testCase.amount, testCase.nonce, testCase.opts...)
			for i := 1; i < len(pks); i++ {
				part := multisig2.Spend(uint8(i), pks[i], principal, testCase.to, testCase.amount, testCase.nonce, testCase.opts...)
				agg.Add(*part.Part(uint8(i)))
			}
			rawTx := agg.Raw()

			decodedTx, err := transactionparser.Parse(rawTx)
			require.NoError(t, err)
			require.Equal(t, testCase.gasPrice, decodedTx.Tx.GetGasPrice())
			require.Equal(t, testCase.to.String(), decodedTx.Tx.GetReceiver().String())
			require.Equal(t, testCase.amount, decodedTx.Tx.GetAmount())
			require.Equal(t, principal.String(), decodedTx.Tx.GetPrincipal().String())
			require.Equal(t, transaction.TypeMultisigSpend, decodedTx.Type)

		})
	}
}

func TestDrainVault(t *testing.T) {
	pubs := make([][]byte, 0, 3)
	pks := make([]ed25519.PrivateKey, 0, 3)
	for i := 0; i < 3; i++ {
		pub, pk, err := ed25519.GenerateKey(rand.New(rand.NewSource(time.Now().UnixNano())))
		require.NoError(t, err)
		pubs = append(pubs, pub)
		pks = append(pks, pk)
	}
	principal := multisig2.Address(multisig.TemplateAddress, 3, pubs...)
	to := types.GenerateAddress(generatePublicKey())
	vaultAddr := types.GenerateAddress(generatePublicKey())
	agg := vesting.DrainVault(
		0,
		pks[0],
		principal,
		vaultAddr,
		to,
		100,
		types.Nonce(1))
	for i := 1; i < len(pks); i++ {
		part := vesting.DrainVault(uint8(i), pks[i], principal, vaultAddr, to, 100, types.Nonce(1))
		agg.Add(*part.Part(uint8(i)))
	}
	txRaw := agg.Raw()

	decodedTx, err := transactionparser.Parse(txRaw)
	require.NoError(t, err)
	require.Equal(t, transaction.TypeDrainVault, decodedTx.Type)
	require.Equal(t, vaultAddr.String(), decodedTx.Vault.GetVault().String())
}

func TestVaultSpawn(t *testing.T) {
	pubs := make([][]byte, 0, 3)
	pks := make([]ed25519.PrivateKey, 0, 3)
	for i := 0; i < 3; i++ {
		pub, pk, err := ed25519.GenerateKey(rand.New(rand.NewSource(time.Now().UnixNano())))
		require.NoError(t, err)
		pubs = append(pubs, pub)
		pks = append(pks, pk)
	}
	owner := types.GenerateAddress(generatePublicKey())
	vaultArgs := &vault.SpawnArguments{
		Owner:               owner,
		InitialUnlockAmount: uint64(1000),
		TotalAmount:         uint64(1001),
		VestingStart:        105120,
		VestingEnd:          4 * 105120,
	}
	vaultAddr := core.ComputePrincipal(vault.TemplateAddress, vaultArgs)

	var agg *multisig2.Aggregator
	for i := 0; i < len(pks); i++ {
		part := multisig2.Spawn(uint8(i), pks[i], vaultAddr, vault.TemplateAddress, vaultArgs, types.Nonce(0))
		if agg == nil {
			agg = part
		} else {
			agg.Add(*part.Part(uint8(i)))
		}
	}
	rawTx := agg.Raw()
	decodedTx, err := transactionparser.Parse(rawTx)
	require.NoError(t, err)

	require.Equal(t, vaultArgs.Owner.String(), decodedTx.Vault.GetOwner().String())
	require.Equal(t, vaultArgs.InitialUnlockAmount, decodedTx.Vault.GetInitialUnlockAmount())
	require.Equal(t, vaultArgs.TotalAmount, decodedTx.Vault.GetTotalAmount())
	require.Equal(t, vaultArgs.VestingStart, decodedTx.Vault.GetVestingStart())
	require.Equal(t, vaultArgs.VestingEnd, decodedTx.Vault.GetVestingEnd())
	require.Equal(t, transaction.TypeVaultSpawn, decodedTx.Type)

}

func generatePublicKey() []byte {
	signer, _ := signing.NewEdSigner()
	return signer.PublicKey().Bytes()
}
