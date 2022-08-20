package collector

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spacemeshos/explorer-backend/model"
)

func TestTransactions(t *testing.T) {
	t.Parallel()
	txs, err := storageDB.GetTransactions(context.TODO(), &bson.D{})
	require.NoError(t, err)
	require.Equal(t, len(generator.Transactions), len(txs))
	for _, tx := range txs {
		// temporary hack, until storage return data as slice of bson.B not an struct.
		txEncoded, err := json.Marshal(tx.Map())
		require.NoError(t, err)
		var tmpTx model.Transaction
		require.NoError(t, json.Unmarshal(txEncoded, &tmpTx))
		generatedTx, ok := generator.Transactions[tmpTx.Id]
		require.True(t, ok)
		tmpTx.Receiver = strings.ToLower(tmpTx.Receiver)
		tmpTx.Sender = strings.ToLower(tmpTx.Sender)
		generatedTx.Receiver = strings.ToLower(generatedTx.Receiver)
		generatedTx.Sender = strings.ToLower(generatedTx.Sender)
		require.Equal(t, *generatedTx, tmpTx)
	}
}
