package collector

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"testing"

	"github.com/spacemeshos/explorer-backend/model"
)

func TestSmeshers(t *testing.T) {
	t.Parallel()
	smeshers, err := storageDB.GetSmeshers(context.TODO(), &bson.D{})
	require.NoError(t, err)
	require.Equal(t, len(generator.Smeshers), len(smeshers))
	for _, smesher := range smeshers {
		// temporary hack, until storage return data as slice of bson.B, not an struct.
		smesherEncoded, err := json.Marshal(smesher.Map())
		require.NoError(t, err)
		var tmpSmesher model.Smesher
		require.NoError(t, json.Unmarshal(smesherEncoded, &tmpSmesher))
		generatedSmesher, ok := generator.Smeshers[strings.ToLower(tmpSmesher.Id)]
		require.True(t, ok)
		generatedSmesher.Id = strings.ToLower(generatedSmesher.Id)
		size, ok := smesher.Map()["cSize"].(int64)
		require.True(t, ok)
		generatedSmesher.CommitmentSize = uint64(size)
		tmpSmesher.Coinbase = strings.ToLower(tmpSmesher.Coinbase)
		generatedSmesher.Coinbase = strings.ToLower(generatedSmesher.Coinbase)
		require.Equal(t, *generatedSmesher, tmpSmesher)
	}
}
