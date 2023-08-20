package storagereader

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"

	bson "go.mongodb.org/mongo-driver/bson"
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
	skip := int64(0)
	if opts[0].Skip != nil {
		skip = *opts[0].Skip
	}

	pipeline := bson.A{
		bson.D{{Key: "$sort", Value: bson.D{{Key: "number", Value: -1}}}},
		bson.D{{Key: "$skip", Value: skip}},
		bson.D{{Key: "$limit", Value: *opts[0].Limit}},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "rewards"},
					{Key: "let",
						Value: bson.D{
							{Key: "start", Value: "$layerstart"},
							{Key: "end", Value: "$layerend"},
						},
					},
					{Key: "pipeline",
						Value: bson.A{
							bson.D{
								{Key: "$match",
									Value: bson.D{
										{Key: "$expr",
											Value: bson.D{
												{Key: "$and",
													Value: bson.A{
														bson.D{
															{Key: "$gte",
																Value: bson.A{
																	"$layer",
																	"$$start",
																},
															},
														},
														bson.D{
															{Key: "$lte",
																Value: bson.A{
																	"$layer",
																	"$$end",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{Key: "as", Value: "rewardsData"},
				},
			},
		},
		bson.D{
			{Key: "$unwind",
				Value: bson.D{
					{Key: "path", Value: "$rewardsData"},
					{Key: "preserveNullAndEmptyArrays", Value: true},
				},
			},
		},
		bson.D{
			{Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: "$_id"},
					{Key: "epochData", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}},
					{Key: "totalRewards", Value: bson.D{{Key: "$sum", Value: "$rewardsData.total"}}},
					{Key: "totalRewardsCount",
						Value: bson.D{
							{Key: "$sum",
								Value: bson.D{
									{Key: "$cond",
										Value: bson.A{
											bson.D{
												{Key: "$gt",
													Value: bson.A{
														"$rewardsData",
														primitive.Null{},
													},
												},
											},
											1,
											0,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		bson.D{{Key: "$project", Value: bson.D{{Key: "epochData.rewardsData", Value: 0}}}},
		bson.D{
			{Key: "$addFields",
				Value: bson.D{
					{Key: "epochData.stats.current.rewards", Value: "$totalRewards"},
					{Key: "epochData.stats.current.rewardsnumber", Value: "$totalRewardsCount"},
					{Key: "epochData.stats.cumulative.rewards", Value: "$totalRewards"},
					{Key: "epochData.stats.cumulative.rewardsnumber", Value: "$totalRewardsCount"},
					{Key: "totalRewards", Value: "$$REMOVE"},
					{Key: "totalRewardsCount", Value: "$$REMOVE"},
				},
			},
		},
		bson.D{{Key: "$replaceRoot", Value: bson.D{{"newRoot", "$epochData"}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "number", Value: -1}}}},
	}

	if query != nil {
		pipeline = append(bson.A{
			bson.D{{Key: "$match", Value: *query}},
		}, pipeline...)
	}

	cursor, err := s.db.Collection("epochs").Aggregate(ctx, pipeline)
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
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "number", Value: epochNumber}}}},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "rewards"},
					{Key: "let",
						Value: bson.D{
							{Key: "start", Value: "$layerstart"},
							{Key: "end", Value: "$layerend"},
						},
					},
					{Key: "pipeline",
						Value: bson.A{
							bson.D{
								{Key: "$match",
									Value: bson.D{
										{Key: "$expr",
											Value: bson.D{
												{Key: "$and",
													Value: bson.A{
														bson.D{
															{Key: "$gte",
																Value: bson.A{
																	"$layer",
																	"$$start",
																},
															},
														},
														bson.D{
															{Key: "$lte",
																Value: bson.A{
																	"$layer",
																	"$$end",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{Key: "as", Value: "rewardsData"},
				},
			},
		},
		bson.D{
			{Key: "$unwind",
				Value: bson.D{
					{Key: "path", Value: "$rewardsData"},
					{Key: "preserveNullAndEmptyArrays", Value: true},
				},
			},
		},
		bson.D{
			{Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: "$_id"},
					{Key: "epochData", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}},
					{Key: "totalRewards", Value: bson.D{{Key: "$sum", Value: "$rewardsData.total"}}},
					{Key: "totalRewardsCount",
						Value: bson.D{
							{Key: "$sum",
								Value: bson.D{
									{Key: "$cond",
										Value: bson.A{
											bson.D{
												{Key: "$gt",
													Value: bson.A{
														"$rewardsData",
														primitive.Null{},
													},
												},
											},
											1,
											0,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		bson.D{{Key: "$project", Value: bson.D{{Key: "epochData.rewardsData", Value: 0}}}},
		bson.D{
			{Key: "$addFields",
				Value: bson.D{
					{Key: "epochData.stats.current.rewards", Value: "$totalRewards"},
					{Key: "epochData.stats.current.rewardsnumber", Value: "$totalRewardsCount"},
					{Key: "epochData.stats.cumulative.rewards", Value: "$totalRewards"},
					{Key: "epochData.stats.cumulative.rewardsnumber", Value: "$totalRewardsCount"},
					{Key: "totalRewards", Value: "$$REMOVE"},
					{Key: "totalRewardsCount", Value: "$$REMOVE"},
				},
			},
		},
		bson.D{{Key: "$replaceRoot", Value: bson.D{{"newRoot", "$epochData"}}}},
	}
	cursor, err := s.db.Collection("epochs").Aggregate(ctx, pipeline)
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
