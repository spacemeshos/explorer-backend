package storagereader

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"

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

	var smeshers []*model.Smesher
	if err = cursor.All(ctx, &smeshers); err != nil {
		return nil, fmt.Errorf("error decode smeshers: %w", err)
	}

	return smeshers, nil
}

// GetEpochSmeshers returns the smeshers for specific epoch
func (s *Reader) CountEpochSmeshers(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	pipeline := bson.A{
		bson.D{
			{"$match", query},
		},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "smeshers"},
					{"localField", "smesher"},
					{"foreignField", "id"},
					{"as", "joinedData"},
				},
			},
		},
		bson.D{{"$unwind", bson.D{{"path", "$joinedData"}}}},
		bson.D{{"$replaceRoot", bson.D{{"newRoot", "$joinedData"}}}},
		bson.D{
			{"$group",
				bson.D{
					{"_id", primitive.Null{}},
					{"total", bson.D{{"$sum", 1}}},
				},
			},
		},
	}

	cursor, err := s.db.Collection("activations").Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("error get smeshers: %w", err)
	}

	if !cursor.Next(ctx) {
		return 0, nil
	}

	doc := cursor.Current
	return utils.GetAsInt64(doc.Lookup("total")), nil
}

// GetEpochSmeshers returns the smeshers for specific epoch
func (s *Reader) GetEpochSmeshers(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Smesher, error) {
	skip := int64(0)
	limit := int64(0)
	if len(opts) > 0 {
		if opts[0].Skip != nil {
			skip = *opts[0].Skip
		}

		if opts[0].Limit != nil {
			limit = *opts[0].Limit
		}
	}

	pipeline := bson.A{
		bson.D{
			{"$match", query},
		},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "smeshers"},
					{"localField", "smesher"},
					{"foreignField", "id"},
					{"as", "joinedData"},
				},
			},
		},
		bson.D{{"$unwind", bson.D{{"path", "$joinedData"}}}},
		bson.D{{"$addFields", bson.D{{"joinedData.atxLayer", "$layer"}}}},
		bson.D{{"$replaceRoot", bson.D{{"newRoot", "$joinedData"}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "atxLayer", Value: -1}}}},
		bson.D{{Key: "$skip", Value: skip}},
	}

	if limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: limit}})
	}

	cursor, err := s.db.Collection("activations").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error get smeshers: %w", err)
	}

	var smeshers []*model.Smesher
	if err = cursor.All(ctx, &smeshers); err != nil {
		return nil, fmt.Errorf("error decode smeshers: %w", err)
	}

	for _, smesher := range smeshers {
		smesher.Timestamp = s.GetLayerTimestamp(smesher.AtxLayer)
	}

	return smeshers, nil
}

// GetSmesher returns the smesher matching the query.
func (s *Reader) GetSmesher(ctx context.Context, smesherID string) (*model.Smesher, error) {
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "id", Value: smesherID}}}}
	lookupStage := bson.D{
		{Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: "malfeasance_proofs"},
				{Key: "localField", Value: "id"},
				{Key: "foreignField", Value: "smesher"},
				{Key: "as", Value: "proofs"},
			},
		},
	}
	cursor, err := s.db.Collection("smeshers").Aggregate(ctx, mongo.Pipeline{
		matchStage,
		lookupStage,
	})
	if err != nil {
		return nil, fmt.Errorf("error get smesher `%s`: %w", smesherID, err)
	}
	if !cursor.Next(ctx) {
		return nil, nil
	}

	var smesher *model.Smesher
	if err = cursor.Decode(&smesher); err != nil {
		return nil, fmt.Errorf("error decode smesher `%s`: %w", smesherID, err)
	}

	return smesher, nil
}

// CountSmesherRewards returns the number of smesher rewards matching the query.
func (s *Reader) CountSmesherRewards(ctx context.Context, smesherID string) (total, count int64, err error) {
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "smesher", Value: smesherID}}}}
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: ""},
			{Key: "total", Value: bson.D{
				{Key: "$sum", Value: "$total"},
			}},
			{Key: "layerReward", Value: bson.D{
				{Key: "$sum", Value: "$layerReward"},
			}},
			{Key: "count", Value: bson.D{
				{Key: "$sum", Value: 1},
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
