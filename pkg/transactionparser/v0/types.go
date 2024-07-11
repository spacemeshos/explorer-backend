package v0

import (
	"github.com/spacemeshos/address"
	"github.com/spacemeshos/go-scale"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/multisig"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"
	"github.com/spacemeshos/go-spacemesh/hash"

	"github.com/spacemeshos/explorer-backend/pkg/transactionparser/transaction"
)

//go:generate scalegen SpawnTransaction SpawnMultisigTransaction SpendTransaction

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

// ComputePrincipal address as the last 20 bytes from blake3(scale(template || args)).
func ComputePrincipal(template core.Address, args scale.Encodable) address.Address {
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
	return ComputePrincipal(wallet.TemplateAddress, &args)
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

// SpawnMultisigTransaction initial transaction for multisig wallet.
type SpawnMultisigTransaction struct {
	Type      uint8
	Principal address.Address
	Method    uint8
	Template  address.Address
	Payload   SpawnMultisigPayload
}

// SpawnMultisigPayload payload of the multisig spawn transaction.
type SpawnMultisigPayload struct {
	Nonce     core.Nonce
	GasPrice  uint64
	Arguments SpawnMultisigArguments
}

// SpawnMultisigArguments arguments for multisig spawn transaction.
type SpawnMultisigArguments struct {
	Required   uint8
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
	return ComputePrincipal(multisig.TemplateAddress, &args)
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

// SpendTransaction coin transfer transaction. also includes multisig.
type SpendTransaction struct {
	Type      uint8
	Principal address.Address
	Method    uint8
	Payload   SpendPayload
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

// SpawnVaultTransaction initial transaction for vault.
type SpawnVaultTransaction struct {
	Type      uint8
	Principal address.Address
	Method    uint8
	Template  address.Address
	Payload   SpawnVaultPayload
}

// SpawnVaultPayload provides arguments for spawn vault transaction.
type SpawnVaultPayload struct {
	Nonce     core.Nonce
	GasPrice  uint64
	Arguments SpawnVaultArguments
}

// SpawnVaultArguments is the arguments of the spawn vault transaction.
type SpawnVaultArguments struct {
	Owner               core.Address
	TotalAmount         uint64
	InitialUnlockAmount uint64
	VestingStart        core.LayerID
	VestingEnd          core.LayerID
}

// GetType returns type of the transaction.
func (t *SpawnVaultTransaction) GetType() uint8 {
	return transaction.TypeSpawn
}

// GetAmount returns amount of the transaction. Always zero for spawn transaction.
func (t *SpawnVaultTransaction) GetAmount() uint64 {
	return 0
}

// GetCounter returns counter of the transaction. Always zero for spawn transaction.
func (t *SpawnVaultTransaction) GetCounter() uint64 {
	return 0
}

// GetReceiver returns receiver address of the transaction.
func (t *SpawnVaultTransaction) GetReceiver() address.Address {
	return address.Address{}
}

// GetGasPrice returns gas price of the transaction.
func (t *SpawnVaultTransaction) GetGasPrice() uint64 {
	return t.Payload.GasPrice
}

// GetPrincipal returns the principal address who pay for gas for this transaction.
func (t *SpawnVaultTransaction) GetPrincipal() address.Address {
	return t.Principal
}

// GetPublicKeys returns public keys of the transaction.
func (t *SpawnVaultTransaction) GetPublicKeys() [][]byte {
	return [][]byte{}
}

func (t *SpawnVaultTransaction) GetOwner() core.Address {
	return t.Payload.Arguments.Owner
}

func (t *SpawnVaultTransaction) GetTotalAmount() uint64 {
	return t.Payload.Arguments.TotalAmount
}

func (t *SpawnVaultTransaction) GetInitialUnlockAmount() uint64 {
	return t.Payload.Arguments.InitialUnlockAmount
}

func (t *SpawnVaultTransaction) GetVestingStart() core.LayerID {
	return t.Payload.Arguments.VestingStart
}

func (t *SpawnVaultTransaction) GetVestingEnd() core.LayerID {
	return t.Payload.Arguments.VestingEnd
}

func (t *SpawnVaultTransaction) GetVault() core.Address {
	return core.Address{}
}

// DrainVaultTransaction initial transaction for vault.
type DrainVaultTransaction struct {
	Type      uint8
	Principal address.Address
	Method    uint8
	Payload   DrainVaultPayload
}

// DrainVaultPayload provides arguments for drain vault transaction.
type DrainVaultPayload struct {
	Nonce     core.Nonce
	GasPrice  uint64
	Arguments DrainVaultArguments
}

// DrainVaultArguments is the arguments of the drain vault transaction.
type DrainVaultArguments struct {
	Vault core.Address
	SpendArguments
}

// GetType returns type of the transaction.
func (t *DrainVaultTransaction) GetType() uint8 {
	return transaction.TypeDrainVault
}

// GetAmount returns amount of the transaction. Always zero for spawn transaction.
func (t *DrainVaultTransaction) GetAmount() uint64 {
	return t.Payload.Arguments.Amount
}

// GetCounter returns counter of the transaction. Always zero for spawn transaction.
func (t *DrainVaultTransaction) GetCounter() uint64 {
	return t.Payload.Nonce
}

// GetReceiver returns receiver address of the transaction.
func (t *DrainVaultTransaction) GetReceiver() address.Address {
	return t.Payload.Arguments.Destination
}

// GetGasPrice returns gas price of the transaction.
func (t *DrainVaultTransaction) GetGasPrice() uint64 {
	return t.Payload.GasPrice
}

// GetPrincipal returns the principal address who pay for gas for this transaction.
func (t *DrainVaultTransaction) GetPrincipal() address.Address {
	return t.Principal
}

// GetPublicKeys returns public keys of the transaction.
func (t *DrainVaultTransaction) GetPublicKeys() [][]byte {
	return [][]byte{}
}

func (t *DrainVaultTransaction) GetOwner() core.Address {
	return core.Address{}
}

func (t *DrainVaultTransaction) GetTotalAmount() uint64 {
	return 0
}

func (t *DrainVaultTransaction) GetInitialUnlockAmount() uint64 {
	return 0
}

func (t *DrainVaultTransaction) GetVestingStart() core.LayerID {
	return core.LayerID(0)
}

func (t *DrainVaultTransaction) GetVestingEnd() core.LayerID {
	return core.LayerID(0)
}

func (t *DrainVaultTransaction) GetVault() core.Address {
	return t.Payload.Arguments.Vault
}
