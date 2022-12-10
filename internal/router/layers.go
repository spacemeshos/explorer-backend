package router

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/spacemeshos/explorer-backend/model"
)

func (a *AppRouter) layers(ctx *fiber.Ctx) error {
	pageNum, pageSize := a.getPagination(ctx)
	layersList, layersTotal, err := a.appService.GetLayers(ctx.UserContext(), pageNum, pageSize)
	if err != nil {
		return fmt.Errorf("failed to get epoch list: %w", err)
	}
	return ctx.JSON(fiber.Map{"data": layersList, "pagination": a.setPaginationResponse(layersTotal, pageNum, pageSize)})
}

func (a *AppRouter) layer(ctx *fiber.Ctx) error {
	layerID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return fiber.ErrBadRequest
	}
	layer, err := a.appService.GetLayer(ctx.UserContext(), layerID)
	if err != nil {
		return fmt.Errorf("failed to get layer info: %w", err)
	}
	return ctx.JSON(fiber.Map{"data": []*model.Layer{layer}})
}

func (a *AppRouter) layerDetails(ctx *fiber.Ctx) error {
	pageNum, pageSize := a.getPagination(ctx)
	layerID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return fmt.Errorf("wrong layer id: %w", err)
	}
	var (
		response interface{}
		total    int64
	)

	switch ctx.Params("entity") {
	case blocks:
		response, total, err = a.appService.GetLayerBlocks(ctx.UserContext(), layerID, pageNum, pageSize)
	case txs:
		response, total, err = a.appService.GetLayerTransactions(ctx.UserContext(), layerID, pageNum, pageSize)
	case smeshers:
		response, total, err = a.appService.GetLayerSmeshers(ctx.UserContext(), layerID, pageNum, pageSize)
	case rewards:
		response, total, err = a.appService.GetLayerRewards(ctx.UserContext(), layerID, pageNum, pageSize)
	case atxs:
		response, total, err = a.appService.GetLayerActivations(ctx.UserContext(), layerID, pageNum, pageSize)
	default:
		return fiber.NewError(fiber.StatusNotFound, "entity not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get layer entity `%s` list: %w", ctx.Params("entity"), err)
	}
	return ctx.JSON(fiber.Map{"data": response, "pagination": a.setPaginationResponse(total, pageNum, pageSize)})
}
