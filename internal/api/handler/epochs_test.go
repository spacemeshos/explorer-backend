package handler_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/spacemeshos/explorer-backend/model"
)

func TestEpochsHandler(t *testing.T) {
	t.Parallel()
	responseEpochs := make([]model.Epoch, 0, 100)
	pageSize := 10
	for i := 1; i <= 10; i++ {
		url := apiPrefix + fmt.Sprintf("/epochs?pagesize=%d", pageSize)
		if i > 0 {
			url += fmt.Sprintf("&page=%d", i)
		}
		var loopResult epochResp
		res := apiServer.Get(t, url)
		res.RequireOK(t)
		res.RequireUnmarshal(t, &loopResult)
		responseEpochs = append(responseEpochs, loopResult.Data...)
	}
	require.Equal(t, len(generator.Epochs), len(responseEpochs))
	sort.Slice(responseEpochs, func(i, j int) bool {
		return responseEpochs[i].Number < responseEpochs[j].Number
	})
	inserted := make([]model.Epoch, 0, len(generator.Epochs))
	for i := range generator.Epochs {
		generator.Epochs[i].Epoch.Stats.Current.Rewards = responseEpochs[i].Stats.Current.Rewards
		generator.Epochs[i].Epoch.Stats.Cumulative.Rewards = responseEpochs[i].Stats.Cumulative.Rewards
		generator.Epochs[i].Epoch.Stats.Current.RewardsNumber = responseEpochs[i].Stats.Current.RewardsNumber
		generator.Epochs[i].Epoch.Stats.Cumulative.RewardsNumber = responseEpochs[i].Stats.Cumulative.RewardsNumber
		inserted = append(inserted, generator.Epochs[i].Epoch)
	}
	require.Equal(t, responseEpochs, inserted)
}

func TestEpochHandler(t *testing.T) {
	t.Parallel()
	for _, ep := range generator.Epochs {
		res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/epochs/%d", ep.Epoch.Number))
		res.RequireOK(t)
		var loopResult epochResp
		res.RequireOK(t)
		res.RequireUnmarshal(t, &loopResult)
		require.Equal(t, 1, len(loopResult.Data))

		ep.Epoch.Stats.Current.Rewards = loopResult.Data[0].Stats.Current.Rewards
		ep.Epoch.Stats.Cumulative.Rewards = loopResult.Data[0].Stats.Cumulative.Rewards
		ep.Epoch.Stats.Current.RewardsNumber = loopResult.Data[0].Stats.Current.RewardsNumber
		ep.Epoch.Stats.Cumulative.RewardsNumber = loopResult.Data[0].Stats.Cumulative.RewardsNumber
		require.Equal(t, ep.Epoch, loopResult.Data[0])
	}
}

func TestEpochLayersHandler(t *testing.T) {
	t.Parallel()
	for _, ep := range generator.Epochs {
		data := make(map[uint32]model.Layer, len(ep.Layers))
		for _, l := range ep.Layers {
			data[l.Layer.Number] = l.Layer
		}

		res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/epochs/%d/layers", ep.Epoch.Number))
		res.RequireOK(t)
		var loopResult layerResp
		res.RequireUnmarshal(t, &loopResult)
		require.Equal(t, len(ep.Layers), len(loopResult.Data))
		for _, l := range loopResult.Data {
			generatedLayer, ok := data[l.Number]
			generatedLayer.Rewards = l.Rewards
			require.True(t, ok)
			require.Equal(t, generatedLayer, l)
		}
	}
}

func TestEpochTxsHandler(t *testing.T) {
	t.Parallel()
	for _, ep := range generator.Epochs {
		res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/epochs/%d/txs?pagesize=100", ep.Epoch.Number))
		res.RequireOK(t)
		var loopResult transactionResp
		res.RequireUnmarshal(t, &loopResult)
		require.Equal(t, len(ep.Transactions), len(loopResult.Data))
		for _, tx := range loopResult.Data {
			generatedTx, ok := ep.Transactions[tx.Id]
			require.True(t, ok)
			require.Equal(t, generatedTx, &tx)
		}
	}
}

func TestEpochSmeshersHandler(t *testing.T) {
	t.Parallel()
	for _, ep := range generator.Epochs {
		res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/epochs/%d/smeshers?pagesize=1000", ep.Epoch.Number))
		res.RequireOK(t)
		var loopResult smesherResp
		res.RequireUnmarshal(t, &loopResult)
		require.Equal(t, len(ep.Smeshers), len(loopResult.Data))
		for _, smesher := range loopResult.Data {
			generatedSmesher, ok := ep.Smeshers[strings.ToLower(smesher.Id)]
			require.True(t, ok)
			smesher.Rewards = generatedSmesher.Rewards // this not calculated on list endpoints, simply set as 0.
			require.Equal(t, *generatedSmesher, smesher)
		}
	}
}

func TestEpochRewardsHandler(t *testing.T) {
	t.Parallel()
	for _, ep := range generator.Epochs {
		res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/epochs/%d/rewards?pagesize=100", ep.Epoch.Number))
		res.RequireOK(t)
		var loopResult rewardResp
		res.RequireUnmarshal(t, &loopResult)
		require.Equal(t, len(ep.Rewards), len(loopResult.Data))
		for _, rw := range loopResult.Data {
			generatedRw, ok := ep.Rewards[rw.Smesher]
			require.True(t, ok)
			generatedRw.ID = rw.ID
			require.Equal(t, *generatedRw, rw)
		}
	}
}

func TestEpochAtxsHandler(t *testing.T) {
	t.Parallel()
	for _, ep := range generator.Epochs {
		res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/epochs/%d/atxs?pagesize=100", ep.Epoch.Number))
		res.RequireOK(t)
		var loopResult atxResp
		res.RequireUnmarshal(t, &loopResult)
		require.Equal(t, len(ep.Activations), len(loopResult.Data))
		for _, atx := range loopResult.Data {
			generatedAtx, ok := ep.Activations[atx.Id]
			require.True(t, ok)
			require.Equal(t, *generatedAtx, atx)
		}
	}
}
