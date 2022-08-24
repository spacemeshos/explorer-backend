package storagereader

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// CountLayers returns the number of layers matching the query.
func (s *StorageReader) CountLayers(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	count, err := s.db.Collection("layers").CountDocuments(ctx, query, opts...)
	if err != nil {
		return 0, fmt.Errorf("error count layers: %w", err)
	}
	return count, nil
}

// GetLayers returns the layers matching the query.
func (s *StorageReader) GetLayers(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Layer, error) {
	cursor, err := s.db.Collection("layers").Find(ctx, query, opts...)
	if err != nil {
		return nil, fmt.Errorf("error get layers: %s", err)
	}

	var layers []*model.Layer
	if err = cursor.All(ctx, &layers); err != nil {
		return nil, fmt.Errorf("error decode layers: %s", err)
	}
	return layers, nil
}

// GetLayer returns the layer matching the query.
func (s *StorageReader) GetLayer(ctx context.Context, layerNumber int) (*model.Layer, error) {
	cursor, err := s.db.Collection("layers").Find(ctx, &bson.D{{"number", layerNumber}})
	if err != nil {
		return nil, fmt.Errorf("error get layer `%d`: %w", layerNumber, err)
	}
	if !cursor.Next(ctx) {
		return nil, nil
	}
	var layer *model.Layer
	if err = cursor.Decode(&layer); err != nil {
		return nil, fmt.Errorf("error decode layer `%d`: %w", layerNumber, err)
	}
	return layer, nil
}
