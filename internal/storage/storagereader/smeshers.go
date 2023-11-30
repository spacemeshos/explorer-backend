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
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "activations"},
					{Key: "let", Value: bson.D{{"smesherId", "$id"}}},
					{Key: "pipeline",
						Value: bson.A{
							bson.D{
								{Key: "$match",
									Value: bson.D{
										{Key: "$expr",
											Value: bson.D{
												{Key: "$eq",
													Value: bson.A{
														"$smesher",
														"$$smesherId",
													},
												},
											},
										},
									},
								},
							},
							bson.D{{Key: "$sort", Value: bson.D{{Key: "layer", Value: 1}}}},
							bson.D{{Key: "$limit", Value: 1}},
							bson.D{
								{Key: "$project",
									Value: bson.D{
										{Key: "_id", Value: 0},
										{Key: "layer", Value: 1},
									},
								},
							},
						},
					},
					{Key: "as", Value: "atxLayerRst"},
				},
			},
		},
		bson.D{
			{Key: "$addFields",
				Value: bson.D{
					{Key: "atxLayer",
						Value: bson.D{
							{Key: "$arrayElemAt",
								Value: bson.A{
									"$atxLayerRst.layer",
									0,
								},
							},
						},
					},
				},
			},
		},
		bson.D{{Key: "$project", Value: bson.D{{Key: "atxLayerRst", Value: 0}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "atxLayer", Value: -1}}}},
		bson.D{{Key: "$skip", Value: skip}},
	}
	if query != nil {
		pipeline = append(bson.A{
			bson.D{{Key: "$match", Value: *query}},
		}, pipeline...)
	}

	if limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: limit}})
	}

	cursor, err := s.db.Collection("smeshers").Aggregate(ctx, pipeline)
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
