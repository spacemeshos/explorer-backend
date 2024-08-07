package handler

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/api/storage"
	"github.com/spacemeshos/go-spacemesh/log"
	"net/http"
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

	return c.JSON(http.StatusOK, overview)
}

func OverviewRefresh(c echo.Context) error {
	cc := c.(*ApiContext)

	overview, err := cc.StorageClient.Overview(cc.Storage)
	if err != nil {
		log.Warning("failed to get overview: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if err = cc.Cache.Set(context.Background(), "overview", overview); err != nil {
		log.Warning("failed to cache overview: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
