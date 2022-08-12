package tests

import (
	"fmt"
	"sort"
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
		require.Equal(t, ep.Epoch, loopResult.Data[0])
	}
}

func TestEpochLayersHandler(t *testing.T) {
	t.Parallel()
	for _, ep := range generator.Epochs {
		res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/epochs/%d/layers", ep.Epoch.Number))
		res.RequireOK(t)
		var loopResult layerResp
		res.RequireUnmarshal(t, &loopResult)
		require.Equal(t, len(ep.Layers), len(loopResult.Data))
		sort.Slice(loopResult.Data, func(i, j int) bool {
			return loopResult.Data[i].Number < loopResult.Data[j].Number
		})
		require.Equal(t, ep.Layers, loopResult.Data)
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
			require.Equal(t, ep.Transactions[tx.Id], tx)
		}
	}
}

func TestEpochSmeshersHandler(t *testing.T) {
	t.Parallel()
	for _, ep := range generator.Epochs {
		res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/epochs/%d/smeshers", ep.Epoch.Number))
		res.RequireOK(t)
		var loopResult smesherResp
		res.RequireUnmarshal(t, &loopResult)
		require.Equal(t, len(ep.Smeshers), len(loopResult.Data))
		for _, tx := range loopResult.Data {
			require.Equal(t, ep.Smeshers[tx.Id], tx)
		}
	}
}

func TestEpochRewardsHandler(t *testing.T) {
	t.Parallel()
	for _, ep := range generator.Epochs {
		res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/epochs/%d/rewards", ep.Epoch.Number))
		res.RequireOK(t)
		var loopResult rewardResp
		res.RequireUnmarshal(t, &loopResult)
		require.Equal(t, len(ep.Rewards), len(loopResult.Data))
		for _, rw := range loopResult.Data {
			require.Equal(t, ep.Rewards[rw.Smesher], rw)
		}
	}
}

func TestEpochAtxsHandler(t *testing.T) {
	t.Parallel()
	for _, ep := range generator.Epochs {
		res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/epochs/%d/atxs", ep.Epoch.Number))
		res.RequireOK(t)
		var loopResult atxResp
		res.RequireUnmarshal(t, &loopResult)
		require.Equal(t, len(ep.Activations), len(loopResult.Data))
		for _, atx := range loopResult.Data {
			require.Equal(t, ep.Activations[atx.Id], atx)
		}
	}
}
