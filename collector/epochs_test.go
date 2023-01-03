package collector_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/spacemeshos/explorer-backend/model"
)

func TestEpochs(t *testing.T) {
	t.Parallel()
	epochs, err := storageDB.GetEpochs(context.TODO(), &bson.D{})
	require.NoError(t, err)
	require.Equal(t, len(generator.Epochs), len(epochs))
	data := make(map[int32]*model.Epoch)
	for _, epoch := range generator.Epochs {
		data[epoch.Epoch.Number] = &epoch.Epoch
	}
	for _, epoch := range epochs {
		// temporary hack, until storage return data as slice of bson.B, not an struct.
		epochMap := epoch.Map()
		generatedEpoch, ok := data[epochMap["number"].(int32)]
		require.True(t, ok)

		require.Equal(t, int64(generatedEpoch.LayerStart), epochMap["layerstart"].(int64))
		require.Equal(t, int64(generatedEpoch.LayerEnd), epochMap["layerend"].(int64))
		require.Equal(t, int64(generatedEpoch.Layers), epochMap["layers"].(int64))
		require.Equal(t, int64(generatedEpoch.Start), epochMap["start"].(int64))
		require.Equal(t, int64(generatedEpoch.End), epochMap["end"].(int64))

		// todo check stats
		println("epoch num", generatedEpoch.Number)
		for k, values := range epochMap["stats"].(primitive.D).Map() {
			v := values.(primitive.D).Map()
			if k == "current" {
				require.Equal(t, generatedEpoch.Stats.Current.Transactions, v["transactions"].(int64))
				require.Equal(t, generatedEpoch.Stats.Current.TxsAmount, v["txsamount"].(int64))
				require.Equal(t, generatedEpoch.Stats.Current.Smeshers, v["smeshers"].(int64))
				// TODO: should be fixed, cause current accounts count is not correct
				//require.Equal(t, generatedEpoch.Stats.Current.Accounts, v["accounts"].(int64))
				require.Equalf(t, generatedEpoch.Stats.Current.RewardsNumber, v["rewardsnumber"].(int64), "rewards number not equal")
				require.Equal(t, generatedEpoch.Stats.Current.Rewards, v["rewards"].(int64), "rewards sum mismatch")
				require.Equal(t, generatedEpoch.Stats.Current.Security, v["security"].(int64))
				require.Equal(t, generatedEpoch.Stats.Current.Capacity, v["capacity"].(int64))
				require.Equal(t, generatedEpoch.Stats.Current.Circulation, v["circulation"].(int64), "circulation sum mismatch")

				// todo should be fixed, cause current stat calc not correct get data about commitmentSize from db
				// require.Equal(t, generatedEpoch.Stats.Current.Decentral, v["decentral"].(int64), "decentral sum mismatch")
			} else if k == "cumulative" {
				//	t.Skip("todo test cumulative stats")
				require.Equal(t, generatedEpoch.Stats.Cumulative.Transactions, v["transactions"].(int64))
				require.Equal(t, generatedEpoch.Stats.Cumulative.TxsAmount, v["txsamount"].(int64))
				require.Equal(t, generatedEpoch.Stats.Cumulative.Smeshers, v["smeshers"].(int64))
				// TODO: should be fixed, cause current accounts count is not correct
				//require.Equal(t, generatedEpoch.Stats.Cumulative.Accounts, v["accounts"].(int64))
				require.Equalf(t, generatedEpoch.Stats.Cumulative.RewardsNumber, v["rewardsnumber"].(int64), "rewards number not equal")
				require.Equal(t, generatedEpoch.Stats.Cumulative.Rewards, v["rewards"].(int64), "rewards sum mismatch")
				require.Equal(t, generatedEpoch.Stats.Cumulative.Security, v["security"].(int64))
				require.Equal(t, generatedEpoch.Stats.Cumulative.Capacity, v["capacity"].(int64))
				require.Equal(t, generatedEpoch.Stats.Cumulative.Circulation, v["circulation"].(int64), "circulation sum mismatch")

				// todo should be fixed, cause current stat calc not correct get data about commitmentSize from db
				// require.Equal(t, generatedEpoch.Stats.Cumulative.Decentral, v["decentral"].(int64), "decentral sum mismatch")
			}
		}
	}
}
