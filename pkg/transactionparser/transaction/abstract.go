package transaction

import (
	"github.com/spacemeshos/address"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/multisig"
)

// DecodedTransactioner is an interface for transaction decoded from raw bytes.
type DecodedTransactioner interface {
	GetType() uint8
	GetAmount() uint64
	GetCounter() uint64
	GetReceiver() address.Address
	GetGasPrice() uint64
	GetPrincipal() address.Address
	GetPublicKeys() [][]byte
}

type DecodedSignature interface {
	GetSignature() []byte
	GetSignatures() []multisig.Part
}

type TransactionData struct {
	Tx         DecodedTransactioner
	Sig        *core.Signature
	Signatures *multisig.Signatures
	Vault      DecodedVault
	Type       int
}

type DecodedVault interface {
	GetVault() core.Address
	GetOwner() core.Address
	GetTotalAmount() uint64
	GetInitialUnlockAmount() uint64
	GetVestingStart() core.LayerID
	GetVestingEnd() core.LayerID
}
