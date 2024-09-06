package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/api/cache"
	"github.com/spacemeshos/explorer-backend/api/storage"

	"github.com/spacemeshos/go-spacemesh/log"
)

func Overview(c echo.Context) error {
	cc := c.(*ApiContext)

	if cached, err := cc.Cache.Get(context.Background(), "overview", new(storage.Overview)); err == nil {
		return c.JSON(http.StatusOK, cached)
	}

	overview, err := cc.StorageClient.Overview(cc.Storage)
	if err != nil {
		log.Warning("failed to get overview: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if err = cc.Cache.Set(context.Background(), "overview", overview); err != nil {
		log.Warning("failed to cache overview: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	cache.LastUpdated.WithLabelValues("/overview").SetToCurrentTime()

	return c.JSON(http.StatusOK, overview)
}

func OverviewRefresh(c echo.Context) error {
	cc := c.(*ApiContext)

	go func() {
		overview, err := cc.StorageClient.Overview(cc.Storage)
		if err != nil {
			log.Warning("failed to get overview: %v", err)
			return
		}

		if err = cc.Cache.Set(context.Background(), "overview", overview); err != nil {
			log.Warning("failed to cache overview: %v", err)
			return
		}

		log.Info("overview refreshed")
		cache.LastUpdated.WithLabelValues("/refresh/overview").SetToCurrentTime()
	}()

	return c.NoContent(http.StatusOK)
}
