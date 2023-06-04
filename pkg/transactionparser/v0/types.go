package v0

import (
	"github.com/spacemeshos/address"
	"github.com/spacemeshos/go-scale"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	"github.com/spacemeshos/go-spacemesh/hash"

	"github.com/spacemeshos/explorer-backend/pkg/transactionparser/transaction"
)

//go:generate scalegen SpawnTransaction SpawnMultisigTransaction SpendTransaction

var (
	// TemplateAddress is an address of the Wallet template.
	TemplateAddress address.Address
	// TemplateAddress1 is an address of the 1/N multisig template.
	TemplateAddress1 address.Address
	// TemplateAddress2 is an address of the 2/N multisig template.
	TemplateAddress2 address.Address
	// TemplateAddress3 is an address of the 3/N multisig template.
	TemplateAddress3 address.Address
)

func init() {
	TemplateAddress[len(TemplateAddress)-1] = 1
	TemplateAddress1[len(TemplateAddress1)-1] = 2
	TemplateAddress2[len(TemplateAddress2)-1] = 3
	TemplateAddress3[len(TemplateAddress3)-1] = 4
}

// Signature returns the signature of the transaction.
type Signature [64]byte

// PublicKey is a public key of the transaction.
type PublicKey [32]byte

// EncodeScale implements scale codec interface.
func (h *PublicKey) EncodeScale(e *scale.Encoder) (int, error) {
	return scale.EncodeByteArray(e, h[:])
}

// DecodeScale implements scale codec interface.
func (h *PublicKey) DecodeScale(d *scale.Decoder) (int, error) {
	return scale.DecodeByteArray(d, h[:])
}

func getMultisigTemplate(pkNum int) address.Address {
	switch pkNum {
	case 2:
		return TemplateAddress1
	case 3:
		return TemplateAddress2
	case 4:
		return TemplateAddress3
	default:
		return TemplateAddress
	}
}

// ComputePrincipal address as the last 20 bytes from blake3(scale(template || args)).
func ComputePrincipal(template address.Address, args scale.Encodable) address.Address {
	hasher := hash.New()
	encoder := scale.NewEncoder(hasher)
	_, _ = template.EncodeScale(encoder)
	_, _ = args.EncodeScale(encoder)
	hs := hasher.Sum(nil)
	return address.GenerateAddress(hs[12:])
}

// SpawnTransaction initial transaction for wallet.
type SpawnTransaction struct {
	Type      uint8
	Principal address.Address
	Method    uint8
	Template  address.Address
	Payload   SpawnPayload
	Sign      Signature
}

// SpawnPayload provides arguments for spawn transaction.
type SpawnPayload struct {
	Nonce     core.Nonce
	GasPrice  uint64
	Arguments SpawnArguments
}

// SpawnArguments is the arguments of the spawn transaction.
type SpawnArguments struct {
	PublicKey PublicKey
}

// GetType returns type of the transaction.
func (t *SpawnTransaction) GetType() uint8 {
	return transaction.TypeSpawn
}

// GetAmount returns amount of the transaction. Always zero for spawn transaction.
func (t *SpawnTransaction) GetAmount() uint64 {
	return 0
}

// GetCounter returns counter of the transaction. Always zero for spawn transaction.
func (t *SpawnTransaction) GetCounter() uint64 {
	return 0
}

// GetReceiver returns receiver address of the transaction.
func (t *SpawnTransaction) GetReceiver() address.Address {
	args := SpawnArguments{}
	copy(args.PublicKey[:], t.Payload.Arguments.PublicKey[:])
	return ComputePrincipal(TemplateAddress, &args)
}

// GetGasPrice returns gas price of the transaction.
func (t *SpawnTransaction) GetGasPrice() uint64 {
	return t.Payload.GasPrice
}

// GetPrincipal returns the principal address who pay for gas for this transaction.
func (t *SpawnTransaction) GetPrincipal() address.Address {
	return t.Principal
}

// GetPublicKeys returns public keys of the transaction.
func (t *SpawnTransaction) GetPublicKeys() [][]byte {
	return [][]byte{t.Payload.Arguments.PublicKey[:]}
}

// GetSignature returns signature of the transaction.
func (t *SpawnTransaction) GetSignature() []byte {
	return t.Sign[:]
}

// SpawnMultisigTransaction initial transaction for multisig wallet.
type SpawnMultisigTransaction struct {
	Type      uint8
	Principal address.Address
	Method    uint8
	Template  address.Address
	Payload   SpawnMultisigPayload
	Sign      Signature
}

// SpawnMultisigPayload payload of the multisig spawn transaction.
type SpawnMultisigPayload struct {
	Arguments SpawnMultisigArguments
	GasPrice  uint64
}

// SpawnMultisigArguments arguments for multisig spawn transaction.
type SpawnMultisigArguments struct {
	PublicKeys []PublicKey `scale:"max=10"`
}

// GetType returns type of the transaction.
func (t *SpawnMultisigTransaction) GetType() uint8 {
	return transaction.TypeMultisigSpawn
}

// GetAmount returns amount of the transaction. Always zero for spawn transaction.
func (t *SpawnMultisigTransaction) GetAmount() uint64 {
	return 0
}

// GetCounter returns counter of the transaction. Always zero for spawn transaction.
func (t *SpawnMultisigTransaction) GetCounter() uint64 {
	return 0
}

// GetReceiver returns receiver address of the transaction.
func (t *SpawnMultisigTransaction) GetReceiver() address.Address {
	args := SpawnMultisigArguments{PublicKeys: make([]PublicKey, len(t.Payload.Arguments.PublicKeys))}
	for i := range t.Payload.Arguments.PublicKeys {
		copy(args.PublicKeys[i][:], t.Payload.Arguments.PublicKeys[i][:])
	}
	return ComputePrincipal(getMultisigTemplate(len(t.Payload.Arguments.PublicKeys)), &args)
}

// GetGasPrice returns gas price of the transaction.
func (t *SpawnMultisigTransaction) GetGasPrice() uint64 {
	return t.Payload.GasPrice
}

// GetPrincipal returns the principal address who pay for gas for this transaction.
func (t *SpawnMultisigTransaction) GetPrincipal() address.Address {
	return t.Principal
}

// GetPublicKeys returns all public keys of the multisig transaction.
func (t *SpawnMultisigTransaction) GetPublicKeys() [][]byte {
	result := make([][]byte, 0, len(t.Payload.Arguments.PublicKeys))
	for i := range t.Payload.Arguments.PublicKeys {
		result = append(result, t.Payload.Arguments.PublicKeys[i][:])
	}
	return result
}

// GetSignature returns the signature of the transaction.
func (t *SpawnMultisigTransaction) GetSignature() []byte {
	return t.Sign[:]
}

// SpendTransaction coin transfer transaction. also includes multisig.
type SpendTransaction struct {
	Type      uint8
	Principal address.Address
	Method    uint8
	Payload   SpendPayload
	Sign      Signature
}

// SpendArguments arguments of the spend transaction.
type SpendArguments struct {
	Destination address.Address
	Amount      uint64
}

// SpendPayload payload of the spend transaction.
type SpendPayload struct {
	Nonce     core.Nonce
	GasPrice  uint64
	Arguments SpendArguments
}

// GetType returns transaction type.
func (t *SpendTransaction) GetType() uint8 {
	return transaction.TypeSpend
}

// GetAmount returns the amount of the transaction.
func (t *SpendTransaction) GetAmount() uint64 {
	return t.Payload.Arguments.Amount
}

// GetCounter returns the counter of the transaction.
func (t *SpendTransaction) GetCounter() uint64 {
	return t.Payload.Nonce
}

// GetReceiver returns receiver address.
func (t *SpendTransaction) GetReceiver() address.Address {
	return t.Payload.Arguments.Destination
}

// GetGasPrice returns gas price of the transaction.
func (t *SpendTransaction) GetGasPrice() uint64 {
	return t.Payload.GasPrice
}

// GetPrincipal return address which spend gas.
func (t *SpendTransaction) GetPrincipal() address.Address {
	return t.Principal
}

// GetPublicKeys returns nil.
func (t *SpendTransaction) GetPublicKeys() [][]byte {
	return nil // todo we do not encode publickeys in the transaction
}

// GetSignature returns signature.
func (t *SpendTransaction) GetSignature() []byte {
	return t.Sign[:]
}
