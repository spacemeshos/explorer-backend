package testseed

import "github.com/spacemeshos/explorer-backend/model"

// BlockContainer container for block and related transactions.
type BlockContainer struct {
	Block        *model.Block
	Transactions []*model.Transaction
	SmesherID    string
}

// LayerContainer container for layer and related blocks, activations and smeshers.
type LayerContainer struct {
	Layer       model.Layer
	Blocks      []*BlockContainer
	Activations map[string]*model.Activation
	Smeshers    map[string]*model.Smesher
}

// GetLastLayer return last generated layer.
func (s *SeedGenerator) GetLastLayer() (curLayer, latestLayer, verifiedLayer uint32) {
	for _, epoch := range s.Epochs {
		for _, layer := range epoch.Layers {
			curLayer = layer.Layer.Number
			latestLayer = layer.Layer.Number
			verifiedLayer = layer.Layer.Number
		}
	}
	return
}
