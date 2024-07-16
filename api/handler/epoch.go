package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/api/cache"
	"github.com/spacemeshos/go-spacemesh/log"
	"net/http"
	"strconv"
)

func EpochStats(c echo.Context) error {
	cc := c.(*ApiContext)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if cached, ok := cache.Cache.Get("epochStats" + c.Param("id")); ok {
		return c.JSON(http.StatusOK, cached)
	}

	epochStats, err := cc.StorageClient.GetEpochStats(cc.Storage, int64(id), cc.LayersPerEpoch)
	if err != nil {
		log.Warning("failed to get layer stats: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	cache.Cache.Set("epochStats"+c.Param("id"), epochStats, 0)
	return c.JSON(http.StatusOK, epochStats)
}
