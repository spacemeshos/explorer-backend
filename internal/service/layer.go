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
		layers, err := e.storage.GetLayers(ctx, &bson.D{}, options.Find().SetSort(bson.D{{Key: "number", Value: -1}}).SetLimit(1).SetProjection(bson.D{{Key: "_id", Value: 0}}))
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
		return nil, ErrNotFound
	}
	return layer, nil
}

// GetLayerByHash returns layer by hash.
//func (e *Service) GetLayerByHash(ctx context.Context, layerHash string) (*model.Layer, error) {
//	layers, err := e.storage.GetLayers(ctx, &bson.D{{Key: "hash", Value: layerHash}})
//	if err != nil {
//		return nil, fmt.Errorf("error get layer by hash `%s`: %w", layerHash, err)
//	}
//	if len(layers) == 0 {
//		return nil, ErrNotFound
//	}
//	return layers[0], nil
//}

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
	return e.getTransactions(ctx, &bson.D{{Key: "layer", Value: layerNum}}, e.getFindOptions("counter", page, perPage))
}

// GetLayerSmeshers returns smeshers for layer.
func (e *Service) GetLayerSmeshers(ctx context.Context, layerNum int, page, perPage int64) (smeshers []*model.Smesher, total int64, err error) {
	filter := &bson.D{{Key: "layer", Value: layerNum}}
	return e.getSmeshers(ctx, filter, e.getFindOptions("id", page, perPage).SetProjection(bson.D{
		{Key: "id", Value: 0},
		{Key: "layer", Value: 0},
		{Key: "coinbase", Value: 0},
		{Key: "prevAtx", Value: 0},
		{Key: "cSize", Value: 0},
	}))
}

// GetLayerRewards returns rewards for layer.
func (e *Service) GetLayerRewards(ctx context.Context, layerNum int, page, perPage int64) (rewards []*model.Reward, total int64, err error) {
	opts := e.getFindOptions("layer", page, perPage)
	opts.SetProjection(bson.D{})
	return e.getRewards(ctx, &bson.D{{Key: "layer", Value: layerNum}}, opts)
}

// GetLayerActivations returns activations for layer.
func (e *Service) GetLayerActivations(ctx context.Context, layerNum int, page, perPage int64) (atxs []*model.Activation, total int64, err error) {
	return e.getActivations(ctx, &bson.D{{Key: "layer", Value: layerNum}}, e.getFindOptions("id", page, perPage))
}

// GetLayerBlocks returns blocks for layer.
func (e *Service) GetLayerBlocks(ctx context.Context, layerNum int, page, perPage int64) (blocks []*model.Block, total int64, err error) {
	return e.getBlocks(ctx, &bson.D{{Key: "layer", Value: layerNum}}, e.getFindOptions("id", page, perPage))
}
