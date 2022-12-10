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

func TestRewards(t *testing.T) {
	t.Parallel()
	rewards, err := storageDB.GetRewards(context.TODO(), &bson.D{})
	require.NoError(t, err)
	require.Equal(t, len(generator.Rewards), len(rewards))
	for _, reward := range rewards {
		// temporary hack, until storage return data as slice of bson.B not an struct.
		rewardEncoded, err := json.Marshal(reward.Map())
		require.NoError(t, err)
		var tmpReward model.Reward
		require.NoError(t, json.Unmarshal(rewardEncoded, &tmpReward))
		generatedReward, ok := generator.Rewards[strings.ToLower(tmpReward.Smesher)]
		require.True(t, ok, "reward not found")
		generatedReward.Smesher = strings.ToLower(generatedReward.Smesher)
		tmpReward.Smesher = strings.ToLower(tmpReward.Smesher)
		tmpReward.Coinbase = strings.ToLower(tmpReward.Coinbase)
		generatedReward.Coinbase = strings.ToLower(generatedReward.Coinbase)
		tmpReward.ID = "" // id is internal mongo id. before insert to db we do not know it.
		require.Equal(t, *generatedReward, tmpReward)
	}
}
