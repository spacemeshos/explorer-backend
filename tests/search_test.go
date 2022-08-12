package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) { // /search/{id}
	t.Parallel()
	for _, epoch := range generator.Epochs {
		// transactions
		for _, tx := range epoch.Transactions {
			res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/search/%s", tx.Id))
			res.RequireOK(t)
			var loopResp redirect
			res.RequireUnmarshal(t, &loopResp)
			res = apiServer.Get(t, loopResp.Redirect)
			res.RequireOK(t)
			var txResp transactionResp
			res.RequireUnmarshal(t, &txResp)
			require.Equal(t, 1, len(txResp.Data))
			require.Equal(t, tx, txResp.Data[0])
		}

		// layer
		for _, layer := range epoch.Layers {
			res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/search/%d", layer.Number))
			res.RequireOK(t)
			var loopResp redirect
			res.RequireUnmarshal(t, &loopResp)
			res = apiServer.Get(t, loopResp.Redirect)
			res.RequireOK(t)
			var resp layerResp
			res.RequireUnmarshal(t, &resp)
			require.Equal(t, 1, len(resp.Data))
			require.Equal(t, layer, resp.Data[0])
		}
	}
	// account
	for _, acc := range generator.Accounts {
		res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/search/%s", acc.Account.Address))
		res.RequireOK(t)
		var loopResp redirect
		res.RequireUnmarshal(t, &loopResp)
		res = apiServer.Get(t, loopResp.Redirect)
		res.RequireOK(t)
		var resp accountResp
		res.RequireUnmarshal(t, &resp)
		require.Equal(t, 1, len(resp.Data))
		require.Equal(t, acc.Account, resp.Data[0])
	}
}
