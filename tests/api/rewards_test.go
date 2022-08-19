package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRewards(t *testing.T) { //"/rewards"
	t.Parallel()
	insertedRewards := generator.Epochs.GetRewards()
	res := apiServer.Get(t, apiPrefix+"/rewards?pagesize=1000")
	res.RequireOK(t)
	var resp rewardResp
	res.RequireUnmarshal(t, &resp)
	require.Equal(t, len(insertedRewards), len(resp.Data))
	for _, reward := range resp.Data {
		rw, ok := insertedRewards[reward.Smesher]
		require.True(t, ok)
		require.Equal(t, rw, &reward)
	}
}

func TestReward(t *testing.T) { //"/rewards/{id}"
	t.Parallel()
	type rewardRespWithID struct {
		Data []struct {
			ID            string `json:"_id"`
			Layer         int    `json:"layer"`
			Total         int64  `json:"total"`
			LayerReward   int    `json:"layerReward"`
			LayerComputed int    `json:"layerComputed"`
			Coinbase      string `json:"coinbase"`
			Smesher       string `json:"smesher"`
			Space         int    `json:"space"`
			Timestamp     int    `json:"timestamp"`
		} `json:"data"`
		Pagination pagination `json:"pagination"`
	}
	insertedRewards := generator.Epochs.GetRewards()

	res := apiServer.Get(t, apiPrefix+"/rewards?pagesize=100")
	res.RequireOK(t)
	var resp rewardRespWithID
	res.RequireUnmarshal(t, &resp)
	for _, reward := range resp.Data {
		res := apiServer.Get(t, apiPrefix+"/rewards/"+reward.ID)
		res.RequireOK(t)
		var respLoop rewardResp
		res.RequireUnmarshal(t, &respLoop)
		require.Equal(t, 1, len(respLoop.Data))
		rw, ok := insertedRewards[reward.Smesher]
		require.True(t, ok)
		require.Equal(t, rw, &respLoop.Data[0])
	}
}
