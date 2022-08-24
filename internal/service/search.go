package service

import (
	"context"
	"fmt"
	"strconv"
)

// Search try guess entity to search and find related one.
func (e *Service) Search(ctx context.Context, search string) (string, error) {
	switch len(search) {
	case 42:
		if acc, _ := e.GetAccount(ctx, search); acc != nil {
			return "/accounts/" + search, nil
		}
		if block, _ := e.GetBlock(ctx, search); block != nil {
			return "/blocks/" + search, nil
		}
	case 66:
		if tx, _ := e.GetTransaction(ctx, search); tx != nil {
			return "/txs/" + search, nil
		}
		if atx, _ := e.GetActivation(ctx, search); atx != nil {
			return "/atxs/" + search, nil
		}
		if smesher, _ := e.GetSmesher(ctx, search); smesher != nil {
			return "/smeshers/" + search, nil
		}
		if layer, _ := e.GetLayerByHash(ctx, search); layer != nil {
			return fmt.Sprintf("/smeshers/%d", layer.Number), nil
		}
	default:
		if reward, _ := e.GetReward(ctx, search); reward != nil {
			return "rewards/" + search, nil
		}
		id, err := strconv.Atoi(search)
		if err != nil {
			return "", ErrNotFound
		}
		layer, err := e.GetCurrentLayer(ctx)
		if err != nil {
			return "", fmt.Errorf("error get current layer for search: %w", err)
		}
		epoch, err := e.GetCurrentEpoch(ctx)
		if err != nil {
			return "", fmt.Errorf("error get current epoch for search: %w", err)
		}
		if id > int(epoch.Number) {
			if id <= int(layer.Number) && id > 0 {
				return fmt.Sprintf("/layers/%d", id), nil
			}
		} else if id > 0 {
			return fmt.Sprintf("/epochs/%d", id), nil
		}
	}
	return "", ErrNotFound
}
