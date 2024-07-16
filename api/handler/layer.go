package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/api/cache"
	"github.com/spacemeshos/go-spacemesh/log"
	"net/http"
	"strconv"
)

func LayerStats(c echo.Context) error {
	cc := c.(*ApiContext)
	lid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if cached, ok := cache.Cache.Get("layerStats" + c.Param("id")); ok {
		return c.JSON(http.StatusOK, cached)
	}

	layerStats, err := cc.StorageClient.GetLayerStats(cc.Storage, int64(lid))
	if err != nil {
		log.Warning("failed to get layer stats: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	cache.Cache.Set("layerStats"+c.Param("id"), layerStats, 0)
	return c.JSON(http.StatusOK, layerStats)
}
