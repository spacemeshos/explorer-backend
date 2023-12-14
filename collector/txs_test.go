package collector_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestTransactions(t *testing.T) {
	t.Parallel()
	txs, err := storageDB.GetTransactions(context.TODO(), &bson.D{})
	require.NoError(t, err)
	require.Equal(t, len(generator.Transactions), len(txs))
	for _, tx := range txs {
		require.NoError(t, err)
		generatedTx, ok := generator.Transactions[tx.Id]
		require.True(t, ok)
		tx.Receiver = strings.ToLower(tx.Receiver)
		tx.Sender = strings.ToLower(tx.Sender)
		generatedTx.Receiver = strings.ToLower(generatedTx.Receiver)
		generatedTx.Sender = strings.ToLower(generatedTx.Sender)
		generatedTx.PublicKey = "" // we do not encode it to send tx, omit this.
		generatedTx.Signature = "" // we generate sign on emulation of pb stream.
		tx.Signature = ""          // we generate sign on emulation of pb stream.
		require.Equal(t, *generatedTx, tx)
	}
}
