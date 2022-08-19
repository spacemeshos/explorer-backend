package collector

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"testing"

	"github.com/spacemeshos/explorer-backend/model"
)

func TestLayers(t *testing.T) {
	t.Parallel()
	layers, err := storageDB.GetLayers(context.TODO(), &bson.D{})
	require.NoError(t, err)
	require.Equal(t, len(generator.Layers), len(layers))
	for _, layer := range layers {
		// temporary hack, until storage return data as slice of bson.B, not an struct.
		layerEncoded, err := json.Marshal(layer.Map())
		require.NoError(t, err)
		var tmpLayer model.Layer
		require.NoError(t, json.Unmarshal(layerEncoded, &tmpLayer))
		generatedLayer, ok := generator.Layers[tmpLayer.Number]
		require.True(t, ok)
		tmpLayer.Rewards = generatedLayer.Rewards // todo should fill data from proto api
		tmpLayer.Hash = tmpLayer.Hash[2:]         // contain string like `0x...`, cut 0x
		require.Equal(t, *generatedLayer, tmpLayer)
	}
}
