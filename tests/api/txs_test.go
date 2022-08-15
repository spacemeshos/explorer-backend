package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransactions(t *testing.T) { // /txs
	t.Parallel()
	insertedTxs := generator.Epochs.GetTransactions()
	res := apiServer.Get(t, apiPrefix+"/txs?pagesize=100")
	res.RequireOK(t)
	var resp transactionResp
	res.RequireUnmarshal(t, &resp)
	require.Equal(t, len(insertedTxs), len(resp.Data))
	for _, tx := range resp.Data {
		require.Equal(t, insertedTxs[tx.Id], tx)
	}
}

func TestTransaction(t *testing.T) { // /txs/{id}
	t.Parallel()
	insertedTxs := generator.Epochs.GetTransactions()
	for _, tx := range insertedTxs {
		res := apiServer.Get(t, apiPrefix+"/txs/"+tx.Id)
		res.RequireOK(t)
		var resp transactionResp
		res.RequireUnmarshal(t, &resp)
		require.Equal(t, 1, len(resp.Data))
		require.Equal(t, tx, resp.Data[0])
	}
}
