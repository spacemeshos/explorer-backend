package testseed

import (
	"sort"

	"github.com/spacemeshos/explorer-backend/model"
)

// TestServerSeed ...
type TestServerSeed struct {
	NetID          uint64
	EpochNumLayers uint64
	LayersDuration uint64
}

const testAPIServiceDB = "explorer_test"

// GetServerSeed ...
func GetServerSeed() *TestServerSeed {
	return &TestServerSeed{
		NetID:          123,
		EpochNumLayers: 10,
		LayersDuration: 10,
	}
}

// SeedEpoch generated epoch for tests.
type SeedEpoch struct {
	Epoch        model.Epoch
	Layers       []model.Layer
	Transactions map[string]model.Transaction
	Rewards      map[string]model.Reward
	Blocks       map[string]model.Block
	Smeshers     map[string]model.Smesher
	Activations  map[string]model.Activation
}

// SeedEpochs ...
type SeedEpochs []*SeedEpoch

// GetTransactions ...
func (s SeedEpochs) GetTransactions() map[string]model.Transaction {
	result := make(map[string]model.Transaction, 0)
	for _, epoch := range s {
		for _, tx := range epoch.Transactions {
			result[tx.Id] = tx
		}
	}
	return result
}

// GetLayers ...
func (s SeedEpochs) GetLayers() []model.Layer {
	result := make([]model.Layer, 0)
	for _, epoch := range s {
		result = append(result, epoch.Layers...)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Number > result[j].Number
	})
	return result
}

// GetSmeshers ...
func (s SeedEpochs) GetSmeshers() map[string]model.Smesher {
	result := make(map[string]model.Smesher, 0)
	for _, epoch := range s {
		for _, smesher := range epoch.Smeshers {
			result[smesher.Id] = smesher
		}
	}
	return result
}

// GetActivations ...
func (s SeedEpochs) GetActivations() map[string]model.Activation {
	result := make(map[string]model.Activation, 0)
	for _, epoch := range s {
		for _, activation := range epoch.Activations {
			result[activation.Id] = activation
		}
	}
	return result
}

// GetRewards ...
func (s SeedEpochs) GetRewards() map[string]model.Reward {
	result := make(map[string]model.Reward, 0)
	for _, epoch := range s {
		for _, reward := range epoch.Rewards {
			result[reward.Smesher] = reward
		}
	}
	return result
}
