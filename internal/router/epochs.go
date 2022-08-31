package router

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/spacemeshos/explorer-backend/model"
)

func (a *AppRouter) epochs(ctx *fiber.Ctx) error {
	pageNum, pageSize := a.getPagination(ctx)
	epochs, epochsTotal, err := a.appService.GetEpochs(ctx.UserContext(), pageNum, pageSize)
	if err != nil {
		return fmt.Errorf("failed to get epoch list: %w", err)
	}
	return ctx.JSON(fiber.Map{"data": epochs, "pagination": a.setPaginationResponse(epochsTotal, pageNum, pageSize)})
}

func (a *AppRouter) epoch(ctx *fiber.Ctx) error {
	layerNum, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	epochs, err := a.appService.GetEpoch(ctx.UserContext(), layerNum)
	if err != nil {
		return fmt.Errorf("failed to get epoch info: %w", err)
	}
	return ctx.JSON(fiber.Map{"data": []*model.Epoch{epochs}})
}

func (a *AppRouter) epochDetails(ctx *fiber.Ctx) error {
	pageNum, pageSize := a.getPagination(ctx)
	epochID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return fmt.Errorf("wrong epoch id: %w", err)
	}
	var (
		response interface{}
		total    int64
	)

	switch ctx.Params("entity") {
	case layers:
		response, total, err = a.appService.GetEpochLayers(ctx.UserContext(), epochID, pageNum, pageSize)
	case txs:
		response, total, err = a.appService.GetEpochTransactions(ctx.UserContext(), epochID, pageNum, pageSize)
	case smeshers:
		response, total, err = a.appService.GetEpochSmeshers(ctx.UserContext(), epochID, pageNum, pageSize)
	case rewards:
		response, total, err = a.appService.GetEpochRewards(ctx.UserContext(), epochID, pageNum, pageSize)
	case atxs:
		response, total, err = a.appService.GetEpochActivations(ctx.UserContext(), epochID, pageNum, pageSize)
	default:
		return fiber.NewError(fiber.StatusNotFound, "entity not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get epoch entity `%s` list: %w", ctx.Params("entity"), err)
	}
	return ctx.JSON(fiber.Map{"data": response, "pagination": a.setPaginationResponse(total, pageNum, pageSize)})
}
