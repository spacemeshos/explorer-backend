package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/eko/gocache/lib/v4/store"
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/api/cache"
	"github.com/spacemeshos/explorer-backend/api/storage"

	"github.com/spacemeshos/go-spacemesh/log"
)

func Layer(c echo.Context) error {
	cc := c.(*ApiContext)
	lid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if cached, err := cc.Cache.Get(context.Background(), "layerStats"+c.Param("id"),
		new(*storage.LayerStats)); err == nil {
		return c.JSON(http.StatusOK, cached)
	}

	layerStats, err := cc.StorageClient.GetLayerStats(cc.Storage, int64(lid))
	if err != nil {
		log.Warning("failed to get layer stats: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if err = cc.Cache.Set(context.Background(), "layerStats"+c.Param("id"),
		layerStats, store.WithExpiration(cache.ShortExpiration)); err != nil {
		log.Warning("failed to cache layer stats: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	cache.LastUpdated.WithLabelValues("/layer/" + c.Param("id")).SetToCurrentTime()

	return c.JSON(http.StatusOK, layerStats)
}
