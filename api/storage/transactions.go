package storage

import (
	"bytes"
	"fmt"

	spacemeshv2alpha1 "github.com/spacemeshos/api/release/go/spacemesh/v2alpha1"
	"github.com/spacemeshos/go-scale"

	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	"github.com/spacemeshos/go-spacemesh/genvm/registry"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/multisig"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/vault"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/vesting"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"
	"github.com/spacemeshos/go-spacemesh/sql"
)

func (c *Client) GetTransactionsCount(db *sql.Database) (count uint64, err error) {
	_, err = db.Exec(`SELECT COUNT(*)
FROM (
  SELECT distinct id
  FROM transactions
  LEFT JOIN transactions_results_addresses
  ON transactions.id = transactions_results_addresses.tid
);`,
		func(stmt *sql.Statement) {
		},
		func(stmt *sql.Statement) bool {
			count = uint64(stmt.ColumnInt64(0))
			return true
		})
	return
}

func decodeTxArgs(decoder *scale.Decoder) (uint8, *core.Address, scale.Encodable, error) {
	reg := registry.New()
	wallet.Register(reg)
	multisig.Register(reg)
	vesting.Register(reg)
	vault.Register(reg)

	_, _, err := scale.DecodeCompact8(decoder)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("%w: failed to decode version %w", core.ErrMalformed, err)
	}

	var principal core.Address
	if _, err := principal.DecodeScale(decoder); err != nil {
		return 0, nil, nil, fmt.Errorf("%w failed to decode principal: %w", core.ErrMalformed, err)
	}

	method, _, err := scale.DecodeCompact8(decoder)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("%w: failed to decode method selector %w", core.ErrMalformed, err)
	}

	var templateAddress *core.Address
	var handler core.Handler
	switch method {
	case core.MethodSpawn:
		templateAddress = &core.Address{}
		if _, err := templateAddress.DecodeScale(decoder); err != nil {
			return 0, nil, nil, fmt.Errorf("%w failed to decode template address %w", core.ErrMalformed, err)
		}
	case vesting.MethodDrainVault:
		templateAddress = &vesting.TemplateAddress
	default:
		templateAddress = &wallet.TemplateAddress
	}

	handler = reg.Get(*templateAddress)
	if handler == nil {
		return 0, nil, nil, fmt.Errorf("%w: unknown template %s", core.ErrMalformed, *templateAddress)
	}

	var p core.Payload
	if _, err = p.DecodeScale(decoder); err != nil {
		return 0, nil, nil, fmt.Errorf("%w: %w", core.ErrMalformed, err)
	}

	args := handler.Args(method)
	if args == nil {
		return 0, nil, nil, fmt.Errorf("%w: unknown method %s %d", core.ErrMalformed, *templateAddress, method)
	}
	if _, err := args.DecodeScale(decoder); err != nil {
		return 0, nil, nil, fmt.Errorf("%w failed to decode method arguments %w", core.ErrMalformed, err)
	}

	return method, templateAddress, args, nil
}

func toTxContents(rawTx []byte) (*spacemeshv2alpha1.TransactionContents,
	spacemeshv2alpha1.Transaction_TransactionType, error,
) {
	res := &spacemeshv2alpha1.TransactionContents{}
	txType := spacemeshv2alpha1.Transaction_TRANSACTION_TYPE_UNSPECIFIED

	r := bytes.NewReader(rawTx)
	method, template, txArgs, err := decodeTxArgs(scale.NewDecoder(r))
	if err != nil {
		return res, txType, err
	}

	switch method {
	case core.MethodSpawn:
		switch *template {
		case wallet.TemplateAddress:
			args := txArgs.(*wallet.SpawnArguments)
			res.Contents = &spacemeshv2alpha1.TransactionContents_SingleSigSpawn{
				SingleSigSpawn: &spacemeshv2alpha1.ContentsSingleSigSpawn{
					Pubkey: args.PublicKey.String(),
				},
			}
			txType = spacemeshv2alpha1.Transaction_TRANSACTION_TYPE_SINGLE_SIG_SPAWN
		case multisig.TemplateAddress:
			args := txArgs.(*multisig.SpawnArguments)
			contents := &spacemeshv2alpha1.TransactionContents_MultiSigSpawn{
				MultiSigSpawn: &spacemeshv2alpha1.ContentsMultiSigSpawn{
					Required: uint32(args.Required),
				},
			}
			contents.MultiSigSpawn.Pubkey = make([]string, len(args.PublicKeys))
			for i := range args.PublicKeys {
				contents.MultiSigSpawn.Pubkey[i] = args.PublicKeys[i].String()
			}
			res.Contents = contents
			txType = spacemeshv2alpha1.Transaction_TRANSACTION_TYPE_MULTI_SIG_SPAWN
		case vesting.TemplateAddress:
			args := txArgs.(*multisig.SpawnArguments)
			contents := &spacemeshv2alpha1.TransactionContents_VestingSpawn{
				VestingSpawn: &spacemeshv2alpha1.ContentsMultiSigSpawn{
					Required: uint32(args.Required),
				},
			}
			contents.VestingSpawn.Pubkey = make([]string, len(args.PublicKeys))
			for i := range args.PublicKeys {
				contents.VestingSpawn.Pubkey[i] = args.PublicKeys[i].String()
			}
			res.Contents = contents
			txType = spacemeshv2alpha1.Transaction_TRANSACTION_TYPE_VESTING_SPAWN
		case vault.TemplateAddress:
			args := txArgs.(*vault.SpawnArguments)
			res.Contents = &spacemeshv2alpha1.TransactionContents_VaultSpawn{
				VaultSpawn: &spacemeshv2alpha1.ContentsVaultSpawn{
					Owner:               args.Owner.String(),
					TotalAmount:         args.TotalAmount,
					InitialUnlockAmount: args.InitialUnlockAmount,
					VestingStart:        args.VestingStart.Uint32(),
					VestingEnd:          args.VestingEnd.Uint32(),
				},
			}
			txType = spacemeshv2alpha1.Transaction_TRANSACTION_TYPE_VAULT_SPAWN
		}
	case core.MethodSpend:
		args := txArgs.(*wallet.SpendArguments)
		res.Contents = &spacemeshv2alpha1.TransactionContents_Send{
			Send: &spacemeshv2alpha1.ContentsSend{
				Destination: args.Destination.String(),
				Amount:      args.Amount,
			},
		}
		txType = spacemeshv2alpha1.Transaction_TRANSACTION_TYPE_SINGLE_SIG_SEND
		if r.Len() > types.EdSignatureSize {
			txType = spacemeshv2alpha1.Transaction_TRANSACTION_TYPE_MULTI_SIG_SEND
		}
	case vesting.MethodDrainVault:
		args := txArgs.(*vesting.DrainVaultArguments)
		res.Contents = &spacemeshv2alpha1.TransactionContents_DrainVault{
			DrainVault: &spacemeshv2alpha1.ContentsDrainVault{
				Vault:       args.Vault.String(),
				Destination: args.Destination.String(),
				Amount:      args.Amount,
			},
		}
		txType = spacemeshv2alpha1.Transaction_TRANSACTION_TYPE_DRAIN_VAULT
	}

	return res, txType, nil
}
