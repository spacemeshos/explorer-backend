package handler

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func HealthzHandler(c echo.Context) error {
	cc := c.(*ApiContext)
	if err := cc.Service.Ping(context.TODO()); err != nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
	}
	return c.String(http.StatusOK, "OK")
}

func Synced(c echo.Context) error {
	cc := c.(*ApiContext)
	networkInfo, err := cc.Service.GetNetworkInfo(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to check is synced: %w", err)
	}

	if !networkInfo.IsSynced {
		return c.String(http.StatusTooEarly, "SYNCING")
	}

	return c.String(http.StatusOK, "SYNCED")
}

func NetworkInfo(c echo.Context) error {
	cc := c.(*ApiContext)
	networkInfo, epoch, layer, err := cc.Service.GetState(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get current state info: %w", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"network": networkInfo,
		"layer":   layer,
		"epoch":   epoch,
	})
}
