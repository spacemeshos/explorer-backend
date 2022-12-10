package router

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/spacemeshos/explorer-backend/model"
)

func (a *AppRouter) apps(ctx *fiber.Ctx) error {
	pageNum, pageSize := a.getPagination(ctx)
	apps, total, err := a.appService.GetApps(ctx.UserContext(), pageNum, pageSize)
	if err != nil {
		return fmt.Errorf("failed to get apps info: %w", err)
	}
	return ctx.JSON(fiber.Map{"data": apps, "pagination": a.setPaginationResponse(total, pageNum, pageSize)})
}

func (a *AppRouter) app(ctx *fiber.Ctx) error {
	app, err := a.appService.GetApp(ctx.UserContext(), ctx.Params("id"))
	if err != nil {
		return fmt.Errorf("failed to get app `%s` info: %w", ctx.Params("id"), err)
	}
	return ctx.JSON(fiber.Map{"data": []*model.App{app}})
}
