package rest_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActivations(t *testing.T) { // /atxs
	t.Parallel()
	insertedAtxs := generator.Epochs.GetActivations()
	res := apiServer.Get(t, apiPrefix+"/atxs?pagesize=1000")
	res.RequireOK(t)
	var resp atxResp
	res.RequireUnmarshal(t, &resp)
	require.Equal(t, len(insertedAtxs), len(resp.Data))
	for _, atx := range resp.Data {
		generatedAtx, ok := insertedAtxs[atx.Id]
		require.True(t, ok)
		require.Equal(t, generatedAtx, &atx)
	}
}

func TestActivation(t *testing.T) { // /atxs/{id}
	t.Parallel()
	insertedAtxs := generator.Epochs.GetActivations()
	res := apiServer.Get(t, apiPrefix+"/atxs?pagesize=1000")
	res.RequireOK(t)
	var resp atxResp
	res.RequireUnmarshal(t, &resp)
	require.Equal(t, len(insertedAtxs), len(resp.Data))
	for _, atx := range resp.Data {
		response := apiServer.Get(t, apiPrefix+"/atxs/"+atx.Id)
		response.RequireOK(t)
		var respLoop atxResp
		response.RequireUnmarshal(t, &respLoop)
		require.Equal(t, 1, len(respLoop.Data))
		generatedAtx, ok := insertedAtxs[atx.Id]
		require.True(t, ok)
		require.Equal(t, *generatedAtx, respLoop.Data[0])
	}
}
