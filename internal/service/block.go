package service

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// GetBlock returns block by id.
func (e *Service) GetBlock(ctx context.Context, blockID string) (*model.Block, error) {
	blocks, _, err := e.getBlocks(ctx, &bson.D{{Key: "id", Value: strings.ToLower(blockID)}}, options.Find().SetLimit(1).SetProjection(bson.D{{Key: "_id", Value: 0}}))
	if err != nil {
		return nil, fmt.Errorf("error get block by `%s`: %w", blockID, err)
	}
	if len(blocks) == 0 {
		return nil, ErrNotFound
	}
	return blocks[0], nil
}

func (e *Service) getBlocks(ctx context.Context, filter *bson.D, options *options.FindOptions) (blocks []*model.Block, total int64, err error) {
	total, err = e.storage.CountBlocks(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("error count blocks: %w", err)
	}
	if total == 0 {
		return []*model.Block{}, 0, nil
	}
	blocks, err = e.storage.GetBlocks(ctx, filter, options)
	if err != nil {
		return nil, 0, fmt.Errorf("error get blocks: %w", err)
	}
	return blocks, total, nil
}
