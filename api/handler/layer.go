package handler

import (
	"github.com/labstack/echo/v4"
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

	layerStats, err := cc.StorageClient.GetLayerStats(cc.Storage, int64(lid))
	if err != nil {
		log.Warning("failed to get layer stats: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, layerStats)
}
