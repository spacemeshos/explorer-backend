package collector

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"testing"

	"github.com/spacemeshos/explorer-backend/model"
)

func TestBlocks(t *testing.T) {
	t.Parallel()
	blocks, err := storageDB.GetBlocks(context.TODO(), &bson.D{})
	require.NoError(t, err)
	for _, block := range blocks {
		// temporary hack, until storage return data as slice of bson.B, not an struct.
		blockEncoded, err := json.Marshal(block.Map())
		require.NoError(t, err)
		var tmpBlock model.Block
		require.NoError(t, json.Unmarshal(blockEncoded, &tmpBlock))
		generated := generator.Blocks[tmpBlock.Id]
		require.NotNil(t, generated)
		require.Equal(t, *generated, tmpBlock)
	}
}
