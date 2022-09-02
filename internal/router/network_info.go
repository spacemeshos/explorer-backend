package router

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/spacemeshos/go-spacemesh/log"
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

func (a *AppRouter) networkInfoWS(conn *websocket.Conn) {
	if err := a.serveNetworkInfo(conn); err != nil {
		log.Error("error in ws: serve network info: %s", err)
		return
	}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := a.serveNetworkInfo(conn); err != nil {
			log.Error("error in ws: serve network info: %s", err)
			return
		}
	}
}

func (a *AppRouter) serveNetworkInfo(conn *websocket.Conn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	networkInfo, epoch, layer, err := a.appService.GetState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current state info: %w", err)
	}
	if err = conn.WriteJSON(fiber.Map{
		"network": networkInfo,
		"layer":   layer,
		"epoch":   epoch,
	}); err != nil {
		return fmt.Errorf("error in ws: serve network info: %w", err)
	}
	return nil
}
