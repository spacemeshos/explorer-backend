package router

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/spacemeshos/go-spacemesh/log"

	"github.com/spacemeshos/explorer-backend/model"
)

func (a *AppRouter) smeshers(ctx *fiber.Ctx) error {
	pageNum, pageSize := a.getPagination(ctx)
	smeshers, total, err := a.appService.GetSmeshers(ctx.UserContext(), pageNum, pageSize)
	if err != nil {
		log.Error("failed to get smeshers list: %s", err)
		return err
	}
	return ctx.JSON(fiber.Map{"data": smeshers, "pagination": a.setPaginationResponse(total, pageNum, pageSize)})
}

func (a *AppRouter) smesher(ctx *fiber.Ctx) error {
	smesher, err := a.appService.GetSmesher(ctx.UserContext(), ctx.Params("id"))
	if err != nil {
		return fmt.Errorf("failed to get smesher: %w", err)
	}
	return ctx.JSON(fiber.Map{"data": []*model.Smesher{smesher}})
}

func (a *AppRouter) smesherDetails(ctx *fiber.Ctx) error {
	var (
		response interface{}
		err      error
		total    int64
	)
	pageNum, pageSize := a.getPagination(ctx)
	switch ctx.Params("entity") {
	case "atxs":
		response, total, err = a.appService.GetSmesherActivations(ctx.UserContext(), ctx.Params("id"), pageNum, pageSize)
	case "rewards":
		response, total, err = a.appService.GetSmesherRewards(ctx.UserContext(), ctx.Params("id"), pageNum, pageSize)
	default:
		return fiber.NewError(fiber.StatusNotFound, "entity not found")
	}
	if err != nil {
		log.Error("failed to get smesher entity `%s` details: %s", ctx.Params("entity"), err)
		return err
	}
	return ctx.JSON(fiber.Map{"data": response, "pagination": a.setPaginationResponse(total, pageNum, pageSize)})
}
