package collector_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spacemeshos/explorer-backend/model"
)

func TestAtxs(t *testing.T) {
	t.Parallel()
	atxs, err := storageDB.GetActivations(context.TODO(), &bson.D{})
	require.NoError(t, err)
	require.Equal(t, len(generator.Activations), len(atxs))
	for _, atx := range atxs {
		// temporary hack until storage return data as slice of bson.B, not an struct.
		atxEncoded, err := json.Marshal(atx.Map())
		require.NoError(t, err)
		var tmpAtx model.Activation
		require.NoError(t, json.Unmarshal(atxEncoded, &tmpAtx))
		atxGen, ok := generator.Activations[tmpAtx.Id]
		require.True(t, ok)
		tmpAtx.Coinbase = strings.ToLower(tmpAtx.Coinbase)
		atxGen.Coinbase = strings.ToLower(atxGen.Coinbase)
		atxGen.PrevAtx = ""
		require.Equal(t, *atxGen, tmpAtx)
	}
}
