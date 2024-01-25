package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spacemeshos/go-spacemesh/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/internal/storage/storagereader"
	"github.com/spacemeshos/explorer-backend/model"
)

// Service main app service which working with database.
type Service struct {
	networkInfo       *model.NetworkInfo
	networkInfoMU     *sync.RWMutex
	networkInfoLoaded time.Time

	currentEpoch       *model.Epoch
	currentEpochMU     *sync.RWMutex
	currentEpochLoaded time.Time

	currentLayer       *model.Layer
	currentLayerMU     *sync.RWMutex
	currentLayerLoaded time.Time

	cacheTTL time.Duration
	storage  storagereader.StorageReader
}

// NewService creates new service instance.
func NewService(reader storagereader.StorageReader, cacheTTL time.Duration) *Service {
	service := &Service{
		storage:        reader,
		cacheTTL:       cacheTTL,
		networkInfoMU:  &sync.RWMutex{},
		currentEpochMU: &sync.RWMutex{},
		currentLayerMU: &sync.RWMutex{},
	}

	if _, err := service.GetNetworkInfo(context.Background()); err != nil {
		log.Err(fmt.Errorf("error load network info: %w", err))
	}
	return service
}

// GetState returns state of the network, current layer and epoch.
func (e *Service) GetState(ctx context.Context) (*model.NetworkInfo, *model.Epoch, *model.Layer, error) {
	net, err := e.GetNetworkInfo(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get network info: %w", err)
	}
	epoch, err := e.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get epoch: %w", err)
	}
	layer, err := e.GetCurrentLayer(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get layer: %w", err)
	}
	return net, epoch, layer, nil
}

// GetNetworkInfo returns actual network info. Caches data for some time (see networkInfoTTL).
func (e *Service) GetNetworkInfo(ctx context.Context) (net *model.NetworkInfo, err error) {
	e.networkInfoMU.RLock()
	net = e.networkInfo
	loadTime := e.networkInfoLoaded
	e.networkInfoMU.RUnlock()
	if net == nil || loadTime.Add(e.cacheTTL).Unix() < time.Now().Unix() {
		net, err = e.storage.GetNetworkInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed get networkInfo: %w", err)
		}
		e.networkInfoMU.Lock()
		e.networkInfo = net
		e.networkInfoLoaded = time.Now()
		e.networkInfoMU.Unlock()
	}
	return net, nil
}

func (e *Service) getFindOptions(key string, page, perPage int64) *options.FindOptions {
	return options.Find().
		SetSort(bson.D{{Key: key, Value: -1}}).
		SetLimit(perPage).
		SetSkip((page - 1) * perPage).
		SetProjection(bson.D{{Key: "_id", Value: 0}})
}

func (e *Service) getFindOptionsSort(sort bson.D, page, perPage int64) *options.FindOptions {
	return options.Find().
		SetSort(sort).
		SetLimit(perPage).
		SetSkip((page - 1) * perPage).
		SetProjection(bson.D{{Key: "_id", Value: 0}})
}

func (e *Service) getEpochLayers(epoch int) (uint32, uint32) {
	e.networkInfoMU.RLock()
	net := e.networkInfo
	e.networkInfoMU.RUnlock()
	start := uint32(epoch) * net.EpochNumLayers
	end := start + net.EpochNumLayers - 1
	return start, end
}

// Ping checks if the database is reachable.
func (e *Service) Ping(ctx context.Context) error {
	return e.storage.Ping(ctx)
}
