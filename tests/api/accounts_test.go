package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccounts(t *testing.T) { // accounts
	t.Parallel()
	res := apiServer.Get(t, apiPrefix+"/accounts?pagesize=1000")
	res.RequireOK(t)
	var resp accountResp
	res.RequireUnmarshal(t, &resp)
	require.Equal(t, len(generator.Accounts), len(resp.Data))
	for _, acc := range resp.Data {
		accGenerated, ok := generator.Accounts[strings.ToLower(acc.Address)]
		require.True(t, ok)
		// this not calculated on list endpoints
		acc.LayerTms = accGenerated.Account.LayerTms
		acc.Txs = accGenerated.Account.Txs
		require.Equal(t, accGenerated.Account, acc)
	}
}

func TestAccount(t *testing.T) { // /accounts/{id}
	t.Parallel()
	for _, acc := range generator.Accounts {
		res := apiServer.Get(t, apiPrefix+"/accounts/"+acc.Account.Address)
		res.RequireOK(t)
		var resp accountResp
		res.RequireUnmarshal(t, &resp)
		require.Equal(t, 1, len(resp.Data))
		require.Equal(t, acc.Account, resp.Data[0])
	}
}

func TestAccountTransactions(t *testing.T) { // /accounts/{id}/txs
	t.Parallel()
	for _, acc := range generator.Accounts {
		res := apiServer.Get(t, apiPrefix+"/accounts/"+acc.Account.Address+"/txs?pagesize=1000")
		res.RequireOK(t)
		var resp transactionResp
		res.RequireUnmarshal(t, &resp)
		require.Equal(t, len(acc.Transactions), len(resp.Data))
		if len(acc.Transactions) == 0 {
			continue
		}
		for _, tx := range resp.Data {
			generatedTx, ok := acc.Transactions[tx.Id]
			require.True(t, ok)
			require.Equal(t, *generatedTx, tx)
		}
	}
}

func TestAccountRewards(t *testing.T) { // /accounts/{id}/rewards
	t.Parallel()
	for _, acc := range generator.Accounts {
		res := apiServer.Get(t, apiPrefix+"/accounts/"+acc.Account.Address+"/rewards?pagesize=1000")
		res.RequireOK(t)
		var resp rewardResp
		res.RequireUnmarshal(t, &resp)
		require.Equal(t, len(acc.Rewards), len(resp.Data))
		if len(acc.Rewards) == 0 {
			continue
		}
		for _, rw := range resp.Data {
			generatedRw, ok := acc.Rewards[rw.Smesher]
			require.True(t, ok)
			require.Equal(t, *generatedRw, rw)
		}
	}
}
