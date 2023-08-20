package storagereader

import (
	"context"
	"fmt"
	"github.com/spacemeshos/go-spacemesh/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// CountLayers returns the number of layers matching the query.
func (s *Reader) CountLayers(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	count, err := s.db.Collection("layers").CountDocuments(ctx, query, opts...)
	if err != nil {
		return 0, fmt.Errorf("error count layers: %w", err)
	}
	return count, nil
}

// GetLayers returns the layers matching the query.
func (s *Reader) GetLayers(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Layer, error) {
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
					{Key: "localField", Value: "number"},
					{Key: "foreignField", Value: "layer"},
					{Key: "as", Value: "rewardsData"},
				},
			},
		},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$rewardsData"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		bson.D{
			{Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: "$_id"},
					{Key: "layerData", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}},
					{Key: "rewards", Value: bson.D{{Key: "$sum", Value: "$rewardsData.total"}}},
				},
			},
		},
		bson.D{
			{Key: "$project",
				Value: bson.D{
					{Key: "layerData", Value: 1},
					{Key: "rewards", Value: 1},
				},
			},
		},
		bson.D{
			{Key: "$replaceRoot",
				Value: bson.D{
					{Key: "newRoot",
						Value: bson.D{
							{Key: "$mergeObjects",
								Value: bson.A{
									"$layerData",
									bson.D{{"rewards", "$rewards"}},
								},
							},
						},
					},
				},
			},
		},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "rewardsData", Value: 0},
			{Key: "_id", Value: 0},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "number", Value: -1}}}},
	}

	if query != nil {
		pipeline = append(bson.A{
			bson.D{{Key: "$match", Value: *query}},
		}, pipeline...)
	}

	cursor, err := s.db.Collection("layers").Aggregate(ctx, pipeline)
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
func (s *Reader) GetLayer(ctx context.Context, layerNumber int) (*model.Layer, error) {
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "number", Value: layerNumber}}}},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "rewards"},
					{Key: "localField", Value: "number"},
					{Key: "foreignField", Value: "layer"},
					{Key: "as", Value: "rewardsData"},
				},
			},
		},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$rewardsData"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		bson.D{
			{Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: "$_id"},
					{Key: "layerData", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}},
					{Key: "rewards", Value: bson.D{{Key: "$sum", Value: "$rewardsData.total"}}},
				},
			},
		},
		bson.D{
			{Key: "$project",
				Value: bson.D{
					{Key: "layerData", Value: 1},
					{Key: "rewards", Value: 1},
				},
			},
		},
		bson.D{
			{Key: "$replaceRoot",
				Value: bson.D{
					{Key: "newRoot",
						Value: bson.D{
							{Key: "$mergeObjects",
								Value: bson.A{
									"$layerData",
									bson.D{{"rewards", "$rewards"}},
								},
							},
						},
					},
				},
			},
		},
		bson.D{{Key: "$project", Value: bson.D{{Key: "rewardsData", Value: 0}}}},
	}

	cursor, err := s.db.Collection("layers").Aggregate(ctx, pipeline)
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

func (s *Reader) GetLayerTimestamp(layer uint32) uint32 {
	networkInfo, err := s.GetNetworkInfo(context.TODO())
	if err != nil {
		log.Err(fmt.Errorf("getLayerTimestamp: %w", err))
		return 0
	}

	if layer == 0 {
		return networkInfo.GenesisTime
	}
	return networkInfo.GenesisTime + (layer-1)*networkInfo.LayerDuration
}
