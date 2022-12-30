package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/go-spacemesh/log"
	"net/http"
	"syscall"
	"time"

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

func NetworkInfoWS(c echo.Context) error {
	ws, err := Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Error("NetworkInfoWS: upgrade error: %w\n", err)
		return nil
	}
	defer ws.Close()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		if err := serveNetworkInfo(c, ws); err != nil {
			if !errors.Is(err, syscall.EPIPE) {
				log.Error("NetworkInfoWS: serve network info: %s", err)
				return nil
			}
		}
	}

	return nil
}

func serveNetworkInfo(c echo.Context, ws *websocket.Conn) error {
	cc := c.(*ApiContext)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	networkInfo, epoch, layer, err := cc.Service.GetState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current state info: %w", err)
	}

	if err = ws.WriteJSON(map[string]interface{}{
		"network": networkInfo,
		"layer":   layer,
		"epoch":   epoch,
	}); err != nil {
		return fmt.Errorf("serve network info: %w", err)
	}

	return nil
}
