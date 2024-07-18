package handler

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/api/storage"
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

	if cached, err := cc.Cache.Get(context.Background(), "epochStats"+c.Param("id"), new(*storage.EpochStats)); err == nil {
		return c.JSON(http.StatusOK, cached)
	}

	epochStats, err := cc.StorageClient.GetEpochStats(cc.Storage, int64(id), cc.LayersPerEpoch)
	if err != nil {
		log.Warning("failed to get layer stats: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if err = cc.Cache.Set(context.Background(), "epochStats"+c.Param("id"), epochStats); err != nil {
		log.Warning("failed to cache epoch stats: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, epochStats)
}
