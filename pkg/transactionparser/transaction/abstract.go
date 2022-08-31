package transaction

import (
	"github.com/spacemeshos/address"
)

// DecodedTransactioner is an interface for transaction decoded from raw bytes.
type DecodedTransactioner interface {
	GetType() int
	GetAmount() uint64
	GetCounter() uint64
	GetReceiver() address.Address
	GetGasPrice() uint64
	GetPrincipal() address.Address
	GetPublicKeys() [][]byte
	GetSignature() []byte
}
