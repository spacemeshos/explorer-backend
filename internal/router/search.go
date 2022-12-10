package router

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (a *AppRouter) search(ctx *fiber.Ctx) error {
	search := strings.ToLower(ctx.Params("id"))
	redirectURL, err := a.appService.Search(ctx.UserContext(), search)
	if err != nil {
		return fmt.Errorf("error search `%s`: %w", search, err)
	}
	return ctx.JSON(fiber.Map{"redirect": redirectURL})
}
