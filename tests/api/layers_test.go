package api

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/spacemeshos/explorer-backend/model"
)

func TestLayers(t *testing.T) { // /layers
	t.Parallel()
	var result layerResp
	insertedLayers := generator.Epochs.GetLayers()
	res := apiServer.Get(t, apiPrefix+"/layers?pagesize=100")
	res.RequireOK(t)
	res.RequireUnmarshal(t, &result)
	require.Equal(t, insertedLayers, result.Data)
}

func TestLayer(t *testing.T) { // /layers/{id:[0-9]+}
	t.Parallel()
	insertedLayers := generator.Epochs.GetLayers()
	for _, layer := range insertedLayers {
		res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/layers/%d", layer.Number))
		res.RequireOK(t)
		var loopResp layerResp
		res.RequireUnmarshal(t, &loopResp)
		require.Equal(t, 1, len(loopResp.Data))
	}
}

func TestLayerTxs(t *testing.T) { // /layers/{id:[0-9]+}/txs
	t.Parallel()
	for _, epoch := range generator.Epochs {
		for _, layerContainer := range epoch.Layers {
			res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/layers/%d/txs", layerContainer.Layer.Number))
			res.RequireOK(t)
			var loopResp transactionResp
			res.RequireUnmarshal(t, &loopResp)
			layerTx := make(map[string]model.Transaction, len(epoch.Transactions))
			for _, tx := range epoch.Transactions {
				if tx.Layer != layerContainer.Layer.Number {
					continue
				}
				layerTx[tx.Id] = *tx
			}
			require.Equal(t, len(layerTx), len(loopResp.Data))
			if len(layerTx) == 0 {
				continue
			}
			for _, tx := range loopResp.Data {
				require.Equal(t, layerTx[tx.Id], tx)
			}
		}
	}
}

func TestLayerSmeshers(t *testing.T) { // /layers/{id:[0-9]+}/smeshers
	t.Parallel()
	for _, epoch := range generator.Epochs {
		for _, layerContainer := range epoch.Layers {
			res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/layers/%d/smeshers", layerContainer.Layer.Number))
			res.RequireOK(t)
			var loopResp smesherResp
			res.RequireUnmarshal(t, &loopResp)
			for _, smesher := range loopResp.Data {
				generatedSmesher, ok := epoch.Smeshers[strings.ToLower(smesher.Id)]
				require.True(t, ok)
				require.Equal(t, *generatedSmesher, smesher)
			}
		}
	}
}

func TestLayerBlocks(t *testing.T) { // /layers/{id:[0-9]+}/blocks
	t.Parallel()
	for _, epoch := range generator.Epochs {
		for _, layerContainer := range epoch.Layers {
			res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/layers/%d/blocks", layerContainer.Layer.Number))
			res.RequireOK(t)
			var loopResp blockResp
			res.RequireUnmarshal(t, &loopResp)
			for _, block := range loopResp.Data {
				require.Equal(t, epoch.Blocks[block.Id], &block)
			}
		}
	}
}

func TestLayerRewards(t *testing.T) { // /layers/{id:[0-9]+}/rewards
	t.Parallel()
	for _, epoch := range generator.Epochs {
		for _, layerContainer := range epoch.Layers {
			res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/layers/%d/rewards", layerContainer.Layer.Number))
			res.RequireOK(t)
			var loopResp rewardResp
			res.RequireUnmarshal(t, &loopResp)
			for _, tx := range loopResp.Data {
				require.Equal(t, epoch.Rewards[tx.Smesher], &tx)
			}
		}
	}
}

func TestLayerAtxs(t *testing.T) { // /layers/{id:[0-9]+}/atxs
	t.Parallel()
	for _, epoch := range generator.Epochs {
		for _, layerContainer := range epoch.Layers {
			res := apiServer.Get(t, apiPrefix+fmt.Sprintf("/layers/%d/atxs", layerContainer.Layer.Number))
			res.RequireOK(t)
			var loopResp atxResp
			res.RequireUnmarshal(t, &loopResp)
			for _, tx := range loopResp.Data {
				require.Equal(t, epoch.Activations[tx.Id], &tx)
			}
		}
	}
}
