package router

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/spacemeshos/explorer-backend/model"
)

func (a *AppRouter) activations(ctx *fiber.Ctx) error {
	pageNum, pageSize := a.getPagination(ctx)
	atxs, total, err := a.appService.GetActivations(ctx.UserContext(), pageNum, pageSize)
	if err != nil {
		return fmt.Errorf("failed to get apps info: %w", err)
	}
	return ctx.JSON(fiber.Map{"data": atxs, "pagination": a.setPaginationResponse(total, pageNum, pageSize)})
}

func (a *AppRouter) activation(ctx *fiber.Ctx) error {
	atx, err := a.appService.GetActivation(ctx.UserContext(), ctx.Params("id"))
	if err != nil {
		return fmt.Errorf("failed to get activation %s info: %w", ctx.Params("id"), err)
	}
	return ctx.JSON(fiber.Map{"data": []*model.Activation{atx}})
}
