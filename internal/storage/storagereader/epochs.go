package storagereader

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// CountEpochs returns the number of epochs matching the query.
func (s *Reader) CountEpochs(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	count, err := s.db.Collection("epochs").CountDocuments(ctx, query, opts...)
	if err != nil {
		return 0, fmt.Errorf("error count epochs: %w", err)
	}
	return count, nil
}

// GetEpochs returns the epochs matching the query.
func (s *Reader) GetEpochs(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Epoch, error) {
	cursor, err := s.db.Collection("epochs").Find(ctx, query, opts...)
	if err != nil {
		return nil, fmt.Errorf("error get epochs: %w", err)
	}
	var epochs []*model.Epoch
	if err = cursor.All(ctx, &epochs); err != nil {
		return nil, err
	}
	return epochs, nil
}

// GetEpoch returns the epoch matching the query.
func (s *Reader) GetEpoch(ctx context.Context, epochNumber int) (*model.Epoch, error) {
	cursor, err := s.db.Collection("epochs").Find(ctx, &bson.D{{"number", epochNumber}})
	if err != nil {
		return nil, fmt.Errorf("error get epoch `%d`: %w", epochNumber, err)
	}
	if !cursor.Next(ctx) {
		return nil, nil
	}
	var epoch *model.Epoch
	if err = cursor.Decode(&epoch); err != nil {
		return nil, fmt.Errorf("error decode epoch `%d`: %w", epochNumber, err)
	}
	return epoch, nil
}
