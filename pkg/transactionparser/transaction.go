package transactionparser

import (
	"bytes"
	"fmt"
	"github.com/spacemeshos/go-scale"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	"github.com/spacemeshos/go-spacemesh/genvm/registry"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/multisig"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/vault"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/vesting"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"

	"github.com/spacemeshos/explorer-backend/pkg/transactionparser/transaction"
	v0 "github.com/spacemeshos/explorer-backend/pkg/transactionparser/v0"
)

// Parse parses transaction from raw bytes and returns its type.
func Parse(rawTx []byte) (*transaction.TransactionData, error) {
	version, method, templateAddress, err := decodeHeader(scale.NewDecoder(bytes.NewReader(rawTx)))
	if err != nil {
		return nil, err
	}
	switch version {
	case 0:
		return v0.ParseTransaction(bytes.NewBuffer(rawTx), method, templateAddress)
	default:
		return nil, fmt.Errorf("%w: unsupported version %d", core.ErrMalformed, version)
	}
}

// decodeHeader decodes version, method and template address from *scale.Decoder
func decodeHeader(decoder *scale.Decoder) (uint8, uint8, *core.Address, error) {
	reg := registry.New()
	wallet.Register(reg)
	multisig.Register(reg)
	vesting.Register(reg)
	vault.Register(reg)

	version, _, err := scale.DecodeCompact8(decoder)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("%w: failed to decode version %w", core.ErrMalformed, err)
	}

	var principal core.Address
	if _, err := principal.DecodeScale(decoder); err != nil {
		return 0, 0, nil, fmt.Errorf("%w failed to decode principal: %w", core.ErrMalformed, err)
	}

	method, _, err := scale.DecodeCompact8(decoder)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("%w: failed to decode method selector %w", core.ErrMalformed, err)
	}

	var templateAddress *core.Address
	if method == core.MethodSpawn {
		templateAddress = &core.Address{}
		if _, err := templateAddress.DecodeScale(decoder); err != nil {
			return 0, 0, nil, fmt.Errorf("%w failed to decode template address %w", core.ErrMalformed, err)
		}
	} else {
		templateAddress = &wallet.TemplateAddress
	}

	return version, method, templateAddress, nil
}
