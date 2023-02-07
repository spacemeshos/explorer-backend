package v0

import (
	"fmt"

	"github.com/spacemeshos/go-spacemesh/codec"
	"github.com/spacemeshos/go-spacemesh/genvm/core"

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
func ParseTransaction(rawTx []byte, method uint32) (transaction.DecodedTransactioner, error) {
	switch method {
	case methodSpawn:
		//var spawnMultisigTx SpawnMultisigTransaction
		//if err := codec.Decode(rawTx, &spawnMultisigTx); err == nil {
		//	return &spawnMultisigTx, nil
		//}
		var spawnTx SpawnTransaction
		if err := codec.Decode(rawTx, &spawnTx); err == nil {
			return &spawnTx, nil
		}
	case methodSend:
		var spendTx SpendTransaction
		if err := codec.Decode(rawTx, &spendTx); err == nil {
			return &spendTx, nil
		}
	default:
		return nil, fmt.Errorf("%w: unsupported method %d", core.ErrMalformed, method)
	}
	return nil, fmt.Errorf("failed to decode transaction: %w", core.ErrMalformed)
}
