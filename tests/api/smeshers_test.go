package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSmeshersHandler(t *testing.T) { // /smeshers
	t.Parallel()
	smeshers := generator.Epochs.GetSmeshers()
	res := apiServer.Get(t, apiPrefix+"/smeshers?pagesize=100")
	res.RequireOK(t)
	var resp smesherResp
	res.RequireUnmarshal(t, &resp)
	require.Equal(t, len(smeshers), len(resp.Data))
	for _, smesher := range resp.Data {
		require.Equal(t, smeshers[smesher.Id], smesher)
	}
}

func TestSmesherHandler(t *testing.T) { // /smeshers/{id}
	t.Parallel()
	smeshers := generator.Epochs.GetSmeshers()
	for _, smesher := range smeshers {
		res := apiServer.Get(t, apiPrefix+"/smeshers/"+smesher.Id)
		res.RequireOK(t)
		var resp smesherResp
		res.RequireUnmarshal(t, &resp)
		require.Equal(t, 1, len(resp.Data))
		require.Equal(t, smesher, resp.Data[0])
	}
}

func TestSmesherAtxsHandler(t *testing.T) { // /smeshers/{id}/atxs
	t.Parallel()
	for _, epoch := range generator.Epochs {
		for smesherID := range epoch.Smeshers {
			res := apiServer.Get(t, apiPrefix+"/smeshers/"+smesherID+"/atxs")
			res.RequireOK(t)
			var resp atxResp
			res.RequireUnmarshal(t, &resp)
			require.Equal(t, 1, len(resp.Data))
			require.Equal(t, epoch.Activations[resp.Data[0].Id], resp.Data[0])
		}
	}
}

func TestSmesherRewardsHandler(t *testing.T) { // /smeshers/{id}/rewards
	t.Parallel()
	for _, epoch := range generator.Epochs {
		for smesherID := range epoch.Smeshers {
			res := apiServer.Get(t, apiPrefix+"/smeshers/"+smesherID+"/rewards")
			res.RequireOK(t)
			var resp rewardResp
			res.RequireUnmarshal(t, &resp)
			require.Equal(t, 1, len(resp.Data))
			require.Equal(t, epoch.Rewards[resp.Data[0].Smesher], resp.Data[0])
		}
	}
}
