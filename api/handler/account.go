package handler

import (
	"context"
	"net/http"

	"github.com/eko/gocache/lib/v4/store"
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/api/cache"
	"github.com/spacemeshos/explorer-backend/api/storage"

	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/log"
)

func Account(c echo.Context) error {
	cc := c.(*ApiContext)

	address := c.Param("address")
	addr, err := types.StringToAddress(address)
	if err != nil {
		log.Warning("failed to parse account address: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	if cached, err := cc.Cache.Get(context.Background(), "accountStats"+address,
		new(*storage.AccountStats)); err == nil {
		return c.JSON(http.StatusOK, cached)
	}

	accountStats, err := cc.StorageClient.GetAccountsStats(cc.Storage, addr)
	if err != nil {
		log.Warning("failed to get account stats: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if err = cc.Cache.Set(context.Background(), "accountStats"+address, accountStats,
		store.WithExpiration(cache.ShortExpiration)); err != nil {
		log.Warning("failed to cache account stats: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, accountStats)
}
