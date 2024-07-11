package v0

import (
	"bytes"
	"fmt"
	"github.com/spacemeshos/go-spacemesh/codec"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	"github.com/spacemeshos/go-spacemesh/genvm/registry"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/multisig"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/vault"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/vesting"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"

	"github.com/spacemeshos/explorer-backend/pkg/transactionparser/transaction"
)

const (
	methodSpawn = 0
	methodSend  = 16
)

// ParseTransaction parses a transaction encoded in version 0.
// possible two types of transaction:
// 1. spawn transaction - `&sdk.TxVersion, &principal, &sdk.MethodSpawn, &wallet.TemplateAddress, &wallet.SpawnPayload`
// 2. spend transaction - `&sdk.TxVersion, &principal, &sdk.MethodSpend, &wallet.SpendPayload.
// every transaction can be multisig also.
func ParseTransaction(rawTx *bytes.Buffer, method uint8, template *core.Address) (*transaction.TransactionData, error) {
	reg := registry.New()
	wallet.Register(reg)
	multisig.Register(reg)
	vesting.Register(reg)
	vault.Register(reg)

	txData := &transaction.TransactionData{}
	switch method {
	case methodSpawn:
		switch *template {
		case wallet.TemplateAddress:
			var spawnTx SpawnTransaction
			if _, err := codec.DecodeFrom(rawTx, &spawnTx); err != nil {
				return nil, err
			}

			txData.Tx = &spawnTx
			txData.Type = transaction.TypeSpawn
		case multisig.TemplateAddress:
			var spawnMultisigTx SpawnMultisigTransaction
			if _, err := codec.DecodeFrom(rawTx, &spawnMultisigTx); err != nil {
				return nil, err
			}
			txData.Tx = &spawnMultisigTx
			txData.Type = transaction.TypeMultisigSpawn
		case vault.TemplateAddress:
			var spawnVaultTx SpawnVaultTransaction
			if _, err := codec.DecodeFrom(rawTx, &spawnVaultTx); err != nil {
				return nil, err
			}
			txData.Tx = &spawnVaultTx
			txData.Vault = &spawnVaultTx
			txData.Type = transaction.TypeVaultSpawn
		case vesting.TemplateAddress:
			var spawnMultisigTx SpawnMultisigTransaction
			if _, err := codec.DecodeFrom(rawTx, &spawnMultisigTx); err != nil {
				return nil, err
			}
			txData.Tx = &spawnMultisigTx
			txData.Type = transaction.TypeVestingSpawn
		}
	case methodSend:
		var spendTx SpendTransaction
		if _, err := codec.DecodeFrom(rawTx, &spendTx); err != nil {
			return nil, err
		}
		txData.Tx = &spendTx
		txData.Type = transaction.TypeSpend
	case vesting.MethodDrainVault:
		var drainVaultTx DrainVaultTransaction
		if _, err := codec.DecodeFrom(rawTx, &drainVaultTx); err != nil {
			return nil, err
		}
		txData.Tx = &drainVaultTx
		txData.Vault = &drainVaultTx
		txData.Type = transaction.TypeDrainVault
	default:
		return nil, fmt.Errorf("%w: unsupported method %d", core.ErrMalformed, method)
	}

	// decode signature or signatures
	if rawTx.Len() <= types.EdSignatureSize {
		var sig core.Signature
		if _, err := codec.DecodeFrom(rawTx, &sig); err != nil {
			return nil, err
		}
		txData.Sig = &sig
	} else {
		var signatures multisig.Signatures
		for rawTx.Len() > 0 {
			var part multisig.Part
			if _, err := codec.DecodeFrom(rawTx, &part); err != nil {
				return nil, err
			}
			signatures = append(signatures, part)
		}
		txData.Signatures = &signatures

		if txData.Type == transaction.TypeSpend {
			txData.Type = transaction.TypeMultisigSpend
		}
	}

	return txData, nil
}
