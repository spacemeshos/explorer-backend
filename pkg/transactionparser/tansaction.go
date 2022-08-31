package transactionparser

import (
	"fmt"

	"github.com/spacemeshos/go-scale"
	"github.com/spacemeshos/go-spacemesh/genvm/core"

	"github.com/spacemeshos/explorer-backend/pkg/transactionparser/transaction"
	v0 "github.com/spacemeshos/explorer-backend/pkg/transactionparser/v0"
)

// Parse parses transaction from raw bytes and returns its type.
func Parse(decoder *scale.Decoder, rawTx []byte, method uint32) (transaction.DecodedTransactioner, error) {
	version, _, err := scale.DecodeCompact8(decoder)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode version %s", core.ErrMalformed, err.Error())
	}
	switch version {
	case 0:
		return v0.ParseTransaction(rawTx, method)
	default:
		return nil, fmt.Errorf("%w: unsupported version %d", core.ErrMalformed, version)
	}
}
