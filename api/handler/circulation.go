package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/go-spacemesh/log"

	"github.com/spacemeshos/explorer-backend/api/cache"
	"github.com/spacemeshos/explorer-backend/api/storage"
)

func Circulation(c echo.Context) error {
	cc := c.(*ApiContext)

	if cached, err := cc.Cache.Get(context.Background(), "circulation", new(*storage.Circulation)); err == nil {
		return c.JSON(http.StatusOK, cached)
	}

	circulation, err := cc.StorageClient.GetCirculation(cc.Storage)
	if err != nil {
		log.Warning("failed to get circulation: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if err = cc.Cache.Set(context.Background(), "circulation", circulation); err != nil {
		log.Warning("failed to cache circulation: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	cache.LastUpdated.WithLabelValues("/circulation").SetToCurrentTime()

	return c.JSON(http.StatusOK, circulation)
}

func CirculationRefresh(c echo.Context) error {
	cc := c.(*ApiContext)

	go func() {
		circulation, err := cc.StorageClient.GetCirculation(cc.Storage)
		if err != nil {
			log.Warning("failed to get circulation: %v", err)
			return
		}

		if err = cc.Cache.Set(context.Background(), "circulation", circulation); err != nil {
			log.Warning("failed to cache circulation: %v", err)
			return
		}

		log.Info("circulation refreshed")
		cache.LastUpdated.WithLabelValues("/refresh/circulation").SetToCurrentTime()
	}()

	return c.NoContent(http.StatusOK)
}
