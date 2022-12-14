package handler_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSmeshersHandler(t *testing.T) { // /smeshers
	t.Parallel()
	res := apiServer.Get(t, apiPrefix+"/smeshers?pagesize=1000")
	res.RequireOK(t)
	var resp smesherResp
	res.RequireUnmarshal(t, &resp)
	require.Equal(t, len(generator.Smeshers), len(resp.Data))
	for _, smesher := range resp.Data {
		generatedSmesher, ok := generator.Smeshers[strings.ToLower(smesher.Id)]
		require.True(t, ok)
		smesher.Rewards = generatedSmesher.Rewards // for this endpoint we not calculate extra values, cause not use this field
		require.Equal(t, generatedSmesher, &smesher)
	}
}

func TestSmesherHandler(t *testing.T) { // /smeshers/{id}
	t.Parallel()
	for _, smesher := range generator.Smeshers {
		res := apiServer.Get(t, apiPrefix+"/smeshers/"+smesher.Id)
		res.RequireOK(t)
		var resp smesherResp
		res.RequireUnmarshal(t, &resp)
		require.Equal(t, 1, len(resp.Data))
		require.Equal(t, *smesher, resp.Data[0])
	}
}

func TestSmesherAtxsHandler(t *testing.T) { // /smeshers/{id}/atxs
	t.Parallel()
	for _, epoch := range generator.Epochs {
		for _, smesher := range epoch.Smeshers {
			res := apiServer.Get(t, apiPrefix+"/smeshers/"+smesher.Id+"/atxs")
			res.RequireOK(t)
			var resp atxResp
			res.RequireUnmarshal(t, &resp)
			require.Equal(t, 1, len(resp.Data))
			atx, ok := epoch.Activations[resp.Data[0].Id]
			require.True(t, ok)
			require.Equal(t, *atx, resp.Data[0])
		}
	}
}

func TestSmesherRewardsHandler(t *testing.T) { // /smeshers/{id}/rewards
	t.Parallel()
	for _, epoch := range generator.Epochs {
		for _, smesher := range epoch.Smeshers {
			res := apiServer.Get(t, apiPrefix+"/smeshers/"+smesher.Id+"/rewards")
			res.RequireOK(t)
			var resp rewardResp
			res.RequireUnmarshal(t, &resp)
			require.Equal(t, 1, len(resp.Data))
			rw, ok := generator.Rewards[resp.Data[0].Smesher]
			require.True(t, ok)
			require.Equal(t, *rw, resp.Data[0])
		}
	}
}
