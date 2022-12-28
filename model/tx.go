package model

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-scale"

	"github.com/spacemeshos/explorer-backend/pkg/transactionparser"
	"github.com/spacemeshos/explorer-backend/utils"
)

type Transaction struct {
	Id string `json:"id" bson:"id"` //nolint will fix it later

	Layer      uint32 `json:"layer" bson:"layer"`
	Block      string `json:"block" bson:"block"`
	BlockIndex uint32 `json:"blockIndex" bson:"blockIndex"`
	Index      uint32 `json:"index" bson:"index"` // the index of the tx in the ordered list of txs to be executed by stf in the layer
	State      int    `json:"state" bson:"state"`
	Timestamp  uint32 `json:"timestamp" bson:"timestamp"`

	MaxGas   uint64 `json:"maxGas" bson:"maxGas"`
	GasPrice uint64 `json:"gasPrice" bson:"gasPrice"`
	GasUsed  uint64 `bson:"gasUsed" json:"gasUsed"` // gas units used by the transaction (gas price in tx)
	Fee      uint64 `json:"fee" bson:"fee"`         // transaction fee charged for the transaction

	Amount  uint64 `json:"amount" bson:"amount"`   // amount of coin transferred in this tx by sender
	Counter uint64 `json:"counter" bson:"counter"` // tx counter aka nonce

	Type      int    `json:"type" bson:"type"`
	Signature string `json:"signature" bson:"signature"` // the signature itself
	PublicKey string `json:"pubKey" bson:"pubKey"`       // included in schemes which require signer to provide a public key

	Sender   string `json:"sender" bson:"sender"` // tx originator, should match signer inside Signature
	Receiver string `json:"receiver" bson:"receiver"`
	SvmData  string `json:"svmData" bson:"svmData"` // svm binary data. Decode with svm-codec
}

type TransactionReceipt struct {
	Id      string //nolint will fix it later
	Layer   uint32
	Index   uint32 // the index of the tx in the ordered list of txs to be executed by stf in the layer
	Result  int
	GasUsed uint64 // gas units used by the transaction (gas price in tx)
	Fee     uint64 // transaction fee charged for the transaction
	SvmData string // svm binary data. Decode with svm-codec
}

type TransactionService interface {
	GetTransaction(ctx context.Context, txID string) (*Transaction, error)
	GetTransactions(ctx context.Context, page, perPage int64) (txs []*Transaction, total int64, err error)
}

func NewTransactionReceipt(txReceipt *pb.TransactionReceipt) *TransactionReceipt {
	return &TransactionReceipt{
		Id:      utils.BytesToHex(txReceipt.GetId().GetId()),
		Result:  int(txReceipt.GetResult()),
		GasUsed: txReceipt.GetGasUsed(),
		Fee:     txReceipt.GetFee().GetValue(),
		Layer:   uint32(txReceipt.GetLayer().GetNumber()),
		Index:   txReceipt.GetIndex(),
		SvmData: utils.BytesToHex(txReceipt.GetSvmData()),
	}
}

// NewTransaction try to parse the transaction and return a new Transaction struct.
func NewTransaction(in *pb.Transaction, layer uint32, blockID string, timestamp uint32, blockIndex uint32) (*Transaction, error) {
	txDecoded, err := transactionparser.Parse(scale.NewDecoder(bytes.NewReader(in.GetRaw())), in.GetRaw(), in.Method)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction: %w", err)
	}

	tx := &Transaction{
		Id:         utils.BytesToHex(in.GetId()),
		Sender:     txDecoded.GetPrincipal().String(),
		Amount:     txDecoded.GetAmount(),
		Counter:    txDecoded.GetCounter(),
		Layer:      layer,
		Block:      blockID,
		BlockIndex: blockIndex,
		State:      int(pb.TransactionState_TRANSACTION_STATE_PROCESSED),
		Timestamp:  timestamp,
		MaxGas:     in.GetMaxGas(),
		GasPrice:   txDecoded.GetGasPrice(),
		Fee:        in.GetMaxGas() * txDecoded.GetGasPrice(),
		Type:       int(txDecoded.GetType()),
		Signature:  utils.BytesToHex(txDecoded.GetSignature()),
		Receiver:   txDecoded.GetReceiver().String(),
	}
	keys := make([]string, 0, len(txDecoded.GetPublicKeys()))
	for i := range txDecoded.GetPublicKeys() {
		keys = append(keys, utils.BytesToHex(txDecoded.GetPublicKeys()[i]))
	}
	tx.PublicKey = strings.Join(keys, ",")

	return tx, nil
}

func GetTransactionStateFromResult(txResult int) int {
	switch txResult {
	case int(pb.TransactionReceipt_TRANSACTION_RESULT_EXECUTED):
		return int(pb.TransactionState_TRANSACTION_STATE_PROCESSED)
	case int(pb.TransactionReceipt_TRANSACTION_RESULT_BAD_COUNTER):
		return int(pb.TransactionState_TRANSACTION_STATE_CONFLICTING)
	case int(pb.TransactionReceipt_TRANSACTION_RESULT_RUNTIME_EXCEPTION):
		return int(pb.TransactionState_TRANSACTION_STATE_REJECTED)
	case int(pb.TransactionReceipt_TRANSACTION_RESULT_INSUFFICIENT_GAS):
		return int(pb.TransactionState_TRANSACTION_STATE_INSUFFICIENT_FUNDS)
	case int(pb.TransactionReceipt_TRANSACTION_RESULT_INSUFFICIENT_FUNDS):
		return int(pb.TransactionState_TRANSACTION_STATE_INSUFFICIENT_FUNDS)
	}
	return int(pb.TransactionState_TRANSACTION_STATE_UNSPECIFIED)
}
