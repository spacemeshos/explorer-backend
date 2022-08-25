package storagereader

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
	"github.com/spacemeshos/explorer-backend/utils"
)

// CountSmeshers returns the number of smeshers matching the query.
func (s *Reader) CountSmeshers(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	count, err := s.db.Collection("smeshers").CountDocuments(ctx, query, opts...)
	if err != nil {
		return 0, fmt.Errorf("error count transactions: %w", err)
	}
	return count, nil
}

// GetSmeshers returns the smeshers matching the query.
func (s *Reader) GetSmeshers(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Smesher, error) {
	cursor, err := s.db.Collection("smeshers").Find(ctx, query, opts...)
	if err != nil {
		return nil, fmt.Errorf("error get smeshers: %w", err)
	}

	var txs []*model.Smesher
	if err = cursor.All(ctx, &txs); err != nil {
		return nil, fmt.Errorf("error decode smeshers: %w", err)
	}
	return txs, nil
}

// GetSmesher returns the smesher matching the query.
func (s *Reader) GetSmesher(ctx context.Context, smesherID string) (*model.Smesher, error) {
	cursor, err := s.db.Collection("smeshers").Find(ctx, &bson.D{{"id", smesherID}})
	if err != nil {
		return nil, fmt.Errorf("error get smesher `%s`: %w", smesherID, err)
	}
	if !cursor.Next(ctx) {
		return nil, nil
	}
	var layer *model.Smesher
	if err = cursor.Decode(&layer); err != nil {
		return nil, fmt.Errorf("error decode smesher `%s`: %w", smesherID, err)
	}
	return layer, nil
}

// CountSmesherRewards returns the number of smesher rewards matching the query.
func (s *Reader) CountSmesherRewards(ctx context.Context, smesherID string) (total, count int64, err error) {
	matchStage := bson.D{{"$match", bson.D{{"smesher", smesherID}}}}
	groupStage := bson.D{
		{"$group", bson.D{
			{"_id", ""},
			{"total", bson.D{
				{"$sum", "$total"},
			}},
			{"layerReward", bson.D{
				{"$sum", "$layerReward"},
			}},
			{"count", bson.D{
				{"$sum", 1},
			}},
		}},
	}
	cursor, err := s.db.Collection("rewards").Aggregate(ctx, mongo.Pipeline{
		matchStage,
		groupStage,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("error get smesher rewards: %w", err)
	}
	if !cursor.Next(ctx) {
		return 0, 0, nil
	}
	doc := cursor.Current
	return utils.GetAsInt64(doc.Lookup("total")), utils.GetAsInt64(doc.Lookup("count")), nil
}
