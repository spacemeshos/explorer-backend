package router

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/spacemeshos/explorer-backend/model"
)

func (a *AppRouter) rewards(ctx *fiber.Ctx) error {
	pageNum, pageSize := a.getPagination(ctx)
	rewards, total, err := a.appService.GetRewards(ctx.UserContext(), pageNum, pageSize)
	if err != nil {
		return fmt.Errorf("failed to get rewards info: %w", err)
	}
	return ctx.JSON(fiber.Map{"data": rewards, "pagination": a.setPaginationResponse(total, pageNum, pageSize)})
}

func (a *AppRouter) reward(ctx *fiber.Ctx) error {
	reward, err := a.appService.GetReward(ctx.UserContext(), ctx.Params("id"))
	if err != nil {
		return fmt.Errorf("failed to get reward `%s` info: %w", ctx.Params("id"), err)
	}
	return ctx.JSON(fiber.Map{"data": []*model.Reward{reward}})
}
