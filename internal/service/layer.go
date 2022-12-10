package service

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// GetCurrentLayer returns current layer.
func (e *Service) GetCurrentLayer(ctx context.Context) (*model.Layer, error) {
	e.currentLayerMU.RLock()
	layer := e.currentLayer
	loadTime := e.currentLayerLoaded
	e.currentLayerMU.RUnlock()
	if layer == nil || loadTime.Add(e.cacheTTL).Unix() < time.Now().Unix() {
		layers, err := e.storage.GetLayers(ctx, &bson.D{}, options.Find().SetSort(bson.D{{"number", -1}}).SetLimit(1).SetProjection(bson.D{{"_id", 0}}))
		if err != nil {
			return nil, fmt.Errorf("error get layers: %s", err)
		}
		if len(layers) == 0 {
			return nil, nil
		}
		layer = layers[0]

		e.currentLayerMU.Lock()
		e.currentLayer = layer
		e.currentLayerLoaded = time.Now()
		e.currentLayerMU.Unlock()
	}
	return layer, nil
}

// GetLayer returns layer by number.
func (e *Service) GetLayer(ctx context.Context, layerNum int) (*model.Layer, error) {
	layer, err := e.storage.GetLayer(ctx, layerNum)
	if err != nil {
		return nil, fmt.Errorf("error get layer %d: %w", layerNum, err)
	}
	if layer == nil {
		return nil, fmt.Errorf("layer %d not found: %w", layerNum, ErrNotFound)
	}
	return layer, nil
}

// GetLayerByHash returns layer by hash.
func (e *Service) GetLayerByHash(ctx context.Context, layerHash string) (*model.Layer, error) {
	layers, err := e.storage.GetLayers(ctx, &bson.D{{"hash", layerHash}})
	if err != nil {
		return nil, fmt.Errorf("error get layer by hash `%s`: %w", layerHash, err)
	}
	if len(layers) == 0 {
		return nil, ErrNotFound
	}
	return layers[0], nil
}

// GetLayers returns layers.
func (e *Service) GetLayers(ctx context.Context, page, perPage int64) (layers []*model.Layer, total int64, err error) {
	total, err = e.storage.CountLayers(ctx, &bson.D{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count total layers: %w", err)
	}
	layers, err = e.storage.GetLayers(ctx, &bson.D{}, e.getFindOptions("number", page, perPage))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get layers: %w", err)
	}
	return layers, total, nil
}

// GetLayerTransactions returns transactions for layer.
func (e *Service) GetLayerTransactions(ctx context.Context, layerNum int, page, perPage int64) (txs []*model.Transaction, total int64, err error) {
	return e.getTransactions(ctx, &bson.D{{"layer", layerNum}}, e.getFindOptions("id", page, perPage))
}

// GetLayerSmeshers returns smeshers for layer.
func (e *Service) GetLayerSmeshers(ctx context.Context, layerNum int, page, perPage int64) (smeshers []*model.Smesher, total int64, err error) {
	filter := &bson.D{{"layer", layerNum}}
	return e.getSmeshers(ctx, filter, e.getFindOptions("id", page, perPage).SetProjection(bson.D{
		{"id", 0},
		{"layer", 0},
		{"coinbase", 0},
		{"prevAtx", 0},
		{"cSize", 0},
	}))
}

// GetLayerRewards returns rewards for layer.
func (e *Service) GetLayerRewards(ctx context.Context, layerNum int, page, perPage int64) (rewards []*model.Reward, total int64, err error) {
	return e.getRewards(ctx, &bson.D{{"layer", layerNum}}, e.getFindOptions("smesher", page, perPage))
}

// GetLayerActivations returns activations for layer.
func (e *Service) GetLayerActivations(ctx context.Context, layerNum int, page, perPage int64) (atxs []*model.Activation, total int64, err error) {
	return e.getActivations(ctx, &bson.D{{"layer", layerNum}}, e.getFindOptions("id", page, perPage))
}

// GetLayerBlocks returns blocks for layer.
func (e *Service) GetLayerBlocks(ctx context.Context, layerNum int, page, perPage int64) (blocks []*model.Block, total int64, err error) {
	return e.getBlocks(ctx, &bson.D{{"layer", layerNum}}, e.getFindOptions("id", page, perPage))
}
