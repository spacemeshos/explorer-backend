package model

import (
	"context"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"

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

	GasProvided uint64 `json:"gasProvided" bson:"gasProvided"`
	GasPrice    uint64 `json:"gasPrice" bson:"gasPrice"`
	GasUsed     uint64 `bson:"gasUsed"`        // gas units used by the transaction (gas price in tx)
	Fee         uint64 `json:"fee" bson:"fee"` // transaction fee charged for the transaction

	Amount  uint64 `json:"amount" bson:"amount"`   // amount of coin transferred in this tx by sender
	Counter uint64 `json:"counter" bson:"counter"` // tx counter aka nonce

	Type      int    `json:"type" bson:"type"`
	Scheme    int    `json:"scheme" bson:"scheme"`       // the signature's scheme
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

func NewTransaction(in *pb.Transaction, layer uint32, blockId string, timestamp uint32, blockIndex uint32) *Transaction {
	gas := in.GetGasOffered()
	sig := in.GetSignature()

	tx := &Transaction{
		Id:          utils.BytesToHex(in.GetId().GetId()),
		Sender:      utils.BytesToAddressString(in.GetSender().GetAddress()),
		Amount:      in.GetAmount().GetValue(),
		Counter:     in.GetCounter(),
		Layer:       layer,
		Block:       blockId,
		BlockIndex:  blockIndex,
		State:       int(pb.TransactionState_TRANSACTION_STATE_PROCESSED),
		Timestamp:   timestamp,
		GasProvided: gas.GetGasProvided(),
		GasPrice:    gas.GetGasPrice(),
		Scheme:      int(sig.GetScheme()),
		Signature:   utils.BytesToHex(sig.GetSignature()),
		PublicKey:   utils.BytesToHex(sig.GetPublicKey()),
	}

	if data := in.GetCoinTransfer(); data != nil {
		tx.Receiver = utils.BytesToAddressString(data.GetReceiver().GetAddress())
	} else if data := in.GetSmartContract(); data != nil {
		tx.Type = int(data.GetType())
		tx.SvmData = utils.BytesToHex(data.GetData())
		tx.Receiver = utils.BytesToAddressString(data.GetAccountId().GetAddress())
	}

	return tx
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
