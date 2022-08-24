package router

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func (a *AppRouter) healthzHandler(ctx *fiber.Ctx) error {
	if err := a.appService.Ping(ctx.UserContext()); err != nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
	}
	return ctx.SendString("OK")
}

func (a *AppRouter) synced(ctx *fiber.Ctx) error {
	networkInfo, err := a.appService.GetNetworkInfo(ctx.UserContext())
	if err != nil {
		return fmt.Errorf("failed to check is synced: %w", err)
	}
	status := "SYNCED"
	if !networkInfo.IsSynced {
		ctx.Status(http.StatusTooEarly)
		status = "SYNCING"
	}
	return ctx.SendString(status)
}

func (a *AppRouter) networkInfo(ctx *fiber.Ctx) error {
	networkInfo, epoch, layer, err := a.appService.GetState(ctx.UserContext())
	if err != nil {
		return fmt.Errorf("failed to get current state info: %w", err)
	}
	return ctx.JSON(fiber.Map{
		"network": networkInfo,
		"layer":   layer,
		"epoch":   epoch,
	})
}
