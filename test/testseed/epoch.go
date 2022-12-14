package testseed

import (
	"sort"

	"github.com/spacemeshos/explorer-backend/model"
)

// TestServerSeed test network config for tests.
type TestServerSeed struct {
	GenesisID               []byte
	EpochNumLayers          uint64
	LayersDuration          uint64
	MaxTransactionPerSecond uint64
	GenesisTime             uint64

	BitsPerLabel  uint32
	LabelsPerUnit uint64
	MinNumUnits   uint32
	MaxNumUnits   uint32
}

// GetPostUnitsSize calcluates size of post units.
func (t *TestServerSeed) GetPostUnitsSize() uint64 {
	return (uint64(t.BitsPerLabel) * t.LabelsPerUnit) / 8
}

// GetServerSeed generate test network config.
func GetServerSeed() *TestServerSeed {
	return &TestServerSeed{
		GenesisTime:             1234567,
		GenesisID:               []byte("genesisid"),
		EpochNumLayers:          10,
		LayersDuration:          10,
		MaxTransactionPerSecond: 100,
		BitsPerLabel:            100,
		LabelsPerUnit:           200,
		MinNumUnits:             20,
		MaxNumUnits:             2000,
	}
}

// SeedEpoch generated epoch for tests.
type SeedEpoch struct {
	Epoch              model.Epoch
	Layers             []*LayerContainer
	Transactions       map[string]*model.Transaction
	Rewards            map[string]*model.Reward
	Blocks             map[string]*model.Block
	Smeshers           map[string]*model.Smesher
	SmeshersCommitment map[string]int64
	Activations        map[string]*model.Activation
}

// SeedEpochs wrapper over generated slice of epochs.
type SeedEpochs []*SeedEpoch

// GetTransactions extract all transactions from epochs.
func (s SeedEpochs) GetTransactions() map[string]*model.Transaction {
	result := make(map[string]*model.Transaction, 0)
	for _, epoch := range s {
		for _, tx := range epoch.Transactions {
			result[tx.Id] = tx
		}
	}
	return result
}

// GetLayers extract all layers from epochs.
func (s SeedEpochs) GetLayers() []model.Layer {
	result := make([]model.Layer, 0)
	for _, epoch := range s {
		for _, layer := range epoch.Layers {
			result = append(result, layer.Layer)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Number > result[j].Number
	})
	return result
}

// GetActivations extract all activations from epochs.
func (s SeedEpochs) GetActivations() map[string]*model.Activation {
	result := make(map[string]*model.Activation, 0)
	for _, epoch := range s {
		for _, activation := range epoch.Activations {
			result[activation.Id] = activation
		}
	}
	return result
}

// GetRewards extract all rewards from epochs.
func (s SeedEpochs) GetRewards() map[string]*model.Reward {
	result := make(map[string]*model.Reward, 0)
	for _, epoch := range s {
		for _, reward := range epoch.Rewards {
			result[reward.Smesher] = reward
		}
	}
	return result
}
