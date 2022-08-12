package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActivations(t *testing.T) { // /atxs
	t.Parallel()
	insertedAtxs := generator.Epochs.GetActivations()
	res := apiServer.Get(t, apiPrefix+"/atxs?pagesize=100")
	res.RequireOK(t)
	var resp atxResp
	res.RequireUnmarshal(t, &resp)
	require.Equal(t, len(insertedAtxs), len(resp.Data))
	for _, atx := range resp.Data {
		require.Equal(t, insertedAtxs[atx.Id], atx)
	}
}

func TestActivation(t *testing.T) { // /atxs/{id}
	t.Parallel()
	insertedAtxs := generator.Epochs.GetActivations()
	res := apiServer.Get(t, apiPrefix+"/atxs?pagesize=100")
	res.RequireOK(t)
	var resp atxResp
	for _, atx := range resp.Data {
		res := apiServer.Get(t, apiPrefix+"/atxs/"+atx.Id)
		res.RequireOK(t)
		var respLoop rewardResp
		res.RequireUnmarshal(t, &respLoop)
		require.Equal(t, 1, len(respLoop.Data))
		require.Equal(t, insertedAtxs[atx.Id], respLoop.Data[0])
	}
}
