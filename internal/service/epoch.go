package service

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// GetCurrentEpoch returns current epoch.
func (e *Service) GetCurrentEpoch(ctx context.Context) (*model.Epoch, error) {
	e.currentEpochMU.RLock()
	epoch := e.currentEpoch
	loadTime := e.currentEpochLoaded
	e.currentEpochMU.RUnlock()
	if epoch == nil || loadTime.Add(e.cacheTTL).Unix() < time.Now().Unix() {
		epochs, err := e.storage.GetEpochs(ctx, &bson.D{}, options.Find().SetSort(bson.D{{Key: "number", Value: -1}}).SetLimit(1).SetProjection(bson.D{{Key: "_id", Value: 0}}))
		if err != nil {
			return nil, fmt.Errorf("failed to get epoch: %w", err)
		}
		if len(epochs) == 0 {
			return nil, nil
		}
		epoch = epochs[0]

		e.currentEpochMU.Lock()
		e.currentEpoch = epoch
		e.currentEpochLoaded = time.Now()
		e.currentEpochMU.Unlock()
	}
	return epoch, nil
}

// GetEpoch get epoch by number.
func (e *Service) GetEpoch(ctx context.Context, epochNum int) (*model.Epoch, error) {
	epoch, err := e.storage.GetEpoch(ctx, epochNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get epoch `%d`: %w", epochNum, err)
	}
	if epoch == nil {
		return nil, ErrNotFound
	}
	return epoch, nil
}

// GetEpochs returns list of epochs.
func (e *Service) GetEpochs(ctx context.Context, page, perPage int64) ([]*model.Epoch, int64, error) {
	total, err := e.storage.CountEpochs(ctx, &bson.D{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count total epochs: %w", err)
	}
	epochs, err := e.storage.GetEpochs(ctx, &bson.D{}, e.getFindOptions("number", page, perPage))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get epochs: %w", err)
	}
	return epochs, total, nil
}

// GetEpochLayers returns layers for the given epoch.
func (e *Service) GetEpochLayers(ctx context.Context, epochNum int, page, perPage int64) (layers []*model.Layer, total int64, err error) {
	layerStart, layerEnd := e.getEpochLayers(epochNum)
	filter := &bson.D{{Key: "number", Value: bson.D{{Key: "$gte", Value: layerStart}, {Key: "$lte", Value: layerEnd}}}}
	total, err = e.storage.CountLayers(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count layers for epoch `%d`: %w", epochNum, err)
	}
	if total == 0 {
		return []*model.Layer{}, 0, nil
	}

	layers, err = e.storage.GetLayers(ctx, filter, e.getFindOptions("number", page, perPage))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get layers for epoch `%d`: %w", epochNum, err)
	}

	return layers, total, nil
}

// GetEpochTransactions returns transactions for the given epoch.
func (e *Service) GetEpochTransactions(ctx context.Context, epochNum int, page, perPage int64) (txs []*model.Transaction, total int64, err error) {
	layerStart, layerEnd := e.getEpochLayers(epochNum)
	filter := &bson.D{{Key: "layer", Value: bson.D{{Key: "$gte", Value: layerStart}, {Key: "$lte", Value: layerEnd}}}}
	return e.getTransactions(ctx, filter, e.getFindOptions("id", page, perPage))
}

// GetEpochSmeshers returns smeshers for the given epoch.
func (e *Service) GetEpochSmeshers(ctx context.Context, epochNum int, page, perPage int64) (smeshers []*model.Smesher, total int64, err error) {
	layerStart, layerEnd := e.getEpochLayers(epochNum)
	filter := &bson.D{{Key: "layer", Value: bson.D{{Key: "$gte", Value: layerStart}, {Key: "$lte", Value: layerEnd}}}}
	return e.getSmeshers(ctx, filter, e.getFindOptions("id", page, perPage).SetProjection(bson.D{
		{Key: "id", Value: 0},
		{Key: "layer", Value: 0},
		{Key: "coinbase", Value: 0},
		{Key: "prevAtx", Value: 0},
		{Key: "cSize", Value: 0},
	}))
}

// GetEpochRewards returns rewards for the given epoch.
func (e *Service) GetEpochRewards(ctx context.Context, epochNum int, page, perPage int64) (rewards []*model.Reward, total int64, err error) {
	layerStart, layerEnd := e.getEpochLayers(epochNum)
	filter := &bson.D{{Key: "layer", Value: bson.D{{Key: "$gte", Value: layerStart}, {Key: "$lte", Value: layerEnd}}}}
	opts := e.getFindOptions("layer", page, perPage)
	opts.SetProjection(bson.D{})
	return e.getRewards(ctx, filter, opts)
}

// GetEpochActivations returns activations for the given epoch.
func (e *Service) GetEpochActivations(ctx context.Context, epochNum int, page, perPage int64) (atxs []*model.Activation, total int64, err error) {
	layerStart, layerEnd := e.getEpochLayers(epochNum)
	filter := &bson.D{{Key: "layer", Value: bson.D{{Key: "$gte", Value: layerStart}, {Key: "$lte", Value: layerEnd}}}}
	return e.getActivations(ctx, filter, e.getFindOptions("id", page, perPage))
}
