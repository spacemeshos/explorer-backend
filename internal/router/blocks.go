package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/spacemeshos/go-spacemesh/log"

	"github.com/spacemeshos/explorer-backend/model"
)

func (a *AppRouter) block(ctx *fiber.Ctx) error {
	block, err := a.appService.GetBlock(ctx.UserContext(), ctx.Params("id"))
	if err != nil {
		log.Error("failed to get block `%s` info: %s", block, err)
		return err
	}
	return ctx.JSON(fiber.Map{"data": []*model.Block{block}})
}
