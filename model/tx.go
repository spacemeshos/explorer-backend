package model

import (
	"context"
	"fmt"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/explorer-backend/pkg/transactionparser"
	"github.com/spacemeshos/explorer-backend/pkg/transactionparser/transaction"
	"github.com/spacemeshos/explorer-backend/utils"
	"strings"
)

type Transaction struct {
	Id string `json:"id" bson:"id"` //nolint will fix it later

	Layer      uint32 `json:"layer" bson:"layer"`
	Block      string `json:"block" bson:"block"`
	BlockIndex uint32 `json:"blockIndex" bson:"blockIndex"`
	Index      uint32 `json:"index" bson:"index"` // the index of the tx in the ordered list of txs to be executed by stf in the layer
	State      int    `json:"state" bson:"state"`
	Result     int    `json:"result" bson:"result"`
	Timestamp  uint32 `json:"timestamp" bson:"timestamp"`

	MaxGas   uint64 `json:"maxGas" bson:"maxGas"`
	GasPrice uint64 `json:"gasPrice" bson:"gasPrice"`
	GasUsed  uint64 `bson:"gasUsed" json:"gasUsed"` // gas units used by the transaction (gas price in tx)
	Fee      uint64 `json:"fee" bson:"fee"`         // transaction fee charged for the transaction

	Amount  uint64 `json:"amount" bson:"amount"`   // amount of coin transferred in this tx by sender
	Counter uint64 `json:"counter" bson:"counter"` // tx counter aka nonce

	Type       int             `json:"type" bson:"type"`
	Signature  string          `json:"signature" bson:"signature"`   // the signature itself
	Signatures []SignaturePart `json:"signatures" bson:"signatures"` // the signature itself
	PublicKey  string          `json:"pubKey" bson:"pubKey"`         // included in schemes which require signer to provide a public key

	Sender   string `json:"sender" bson:"sender"` // tx originator, should match signer inside Signature
	Receiver string `json:"receiver" bson:"receiver"`
	SvmData  string `json:"svmData" bson:"svmData"` // svm binary data. Decode with svm-codec

	Message          string   `json:"message" bson:"message"`
	TouchedAddresses []string `json:"touchedAddresses" bson:"touchedAddresses"`

	Vault                    string `json:"vault" bson:"vault"`
	VaultOwner               string `json:"vaultOwner" bson:"vaultOwner"`
	VaultTotalAmount         uint64 `json:"vaultTotalAmount" bson:"vaultTotalAmount"`
	VaultInitialUnlockAmount uint64 `json:"vaultInitialUnlockAmount" bson:"vaultInitialUnlockAmount"`
	VaultVestingStart        uint32 `json:"vaultVestingStart" bson:"vaultVestingStart"`
	VaultVestingEnd          uint32 `json:"vaultVestingEnd" bson:"vaultVestingEnd"`
}

type SignaturePart struct {
	Ref       uint32 `json:"ref" bson:"ref"`
	Signature string `json:"signature" bson:"signature"`
}

type TransactionReceipt struct {
	Id               string //nolint will fix it later
	Result           int
	Message          string
	GasUsed          uint64 // gas units used by the transaction (gas price in tx)
	Fee              uint64 // transaction fee charged for the transaction
	Layer            uint32
	Block            string
	TouchedAddresses []string
}

type TransactionService interface {
	GetTransaction(ctx context.Context, txID string) (*Transaction, error)
	GetTransactions(ctx context.Context, page, perPage int64) (txs []*Transaction, total int64, err error)
}

func NewTransactionResult(res *pb.TransactionResult, state *pb.TransactionState, networkInfo NetworkInfo) (*Transaction, error) {
	layerStart := networkInfo.GenesisTime + res.GetLayer()*networkInfo.LayerDuration
	tx, err := NewTransaction(res.GetTx(), res.GetLayer(), utils.NBytesToHex(res.GetBlock(), 20), layerStart, 0)
	if err != nil {
		return nil, err
	}

	tx.State = int(state.State)
	tx.Fee = res.GetFee()
	tx.GasUsed = res.GetGasConsumed()
	tx.Message = res.GetMessage()
	tx.TouchedAddresses = res.GetTouchedAddresses()
	tx.Result = int(res.Status)

	return tx, nil
}

// NewTransaction try to parse the transaction and return a new Transaction struct.
func NewTransaction(in *pb.Transaction, layer uint32, blockID string, timestamp uint32, blockIndex uint32) (*Transaction, error) {
	txDecoded, err := transactionparser.Parse(in.GetRaw())
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction: %w", err)
	}
	tx := &Transaction{
		Id:         utils.BytesToHex(in.GetId()),
		Sender:     txDecoded.Tx.GetPrincipal().String(),
		Amount:     txDecoded.Tx.GetAmount(),
		Counter:    txDecoded.Tx.GetCounter(),
		Layer:      layer,
		Block:      blockID,
		BlockIndex: blockIndex,
		State:      int(pb.TransactionState_TRANSACTION_STATE_UNSPECIFIED),
		Timestamp:  timestamp,
		MaxGas:     in.GetMaxGas(),
		GasPrice:   txDecoded.Tx.GetGasPrice(),
		Fee:        in.GetMaxGas() * txDecoded.Tx.GetGasPrice(),
		Type:       txDecoded.Type,
		Receiver:   txDecoded.Tx.GetReceiver().String(),
	}
	keys := make([]string, 0, len(txDecoded.Tx.GetPublicKeys()))
	for i := range txDecoded.Tx.GetPublicKeys() {
		keys = append(keys, utils.BytesToHex(txDecoded.Tx.GetPublicKeys()[i]))
	}
	tx.PublicKey = strings.Join(keys, ",")

	if txDecoded.Signatures != nil {
		for _, sig := range *txDecoded.Signatures {
			tx.Signatures = append(tx.Signatures, SignaturePart{
				Ref:       uint32(sig.Ref),
				Signature: utils.BytesToHex(sig.Sig.Bytes()),
			})
		}
	} else {
		tx.Signature = utils.BytesToHex(txDecoded.Sig.Bytes())
	}

	if txDecoded.Type == transaction.TypeDrainVault {
		tx.Vault = txDecoded.Vault.GetVault().String()
	}

	if txDecoded.Type == transaction.TypeVaultSpawn {
		tx.VaultOwner = txDecoded.Vault.GetOwner().String()
		tx.VaultTotalAmount = txDecoded.Vault.GetTotalAmount()
		tx.VaultInitialUnlockAmount = txDecoded.Vault.GetInitialUnlockAmount()
		tx.VaultVestingStart = txDecoded.Vault.GetVestingStart().Uint32()
		tx.VaultVestingEnd = txDecoded.Vault.GetVestingEnd().Uint32()
	}

	return tx, nil
}
