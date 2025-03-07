package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/eko/gocache/lib/v4/store"
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/log"

	"github.com/spacemeshos/explorer-backend/api/cache"
	"github.com/spacemeshos/explorer-backend/api/storage"
)

func Smeshers(c echo.Context) error {
	cc := c.(*ApiContext)
	limit, offset := GetPagination(c)

	if cached, err := cc.Cache.Get(context.Background(), fmt.Sprintf("smeshers-%d-%d", limit, offset),
		new(*storage.SmesherList)); err == nil {
		return c.JSON(http.StatusOK, cached)
	}

	smeshers, err := cc.StorageClient.GetSmeshers(cc.Storage, uint64(limit), uint64(offset))
	if err != nil {
		log.Warning("failed to get smeshers: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if err = cc.Cache.Set(context.Background(), fmt.Sprintf("smeshers-%d-%d", limit, offset), smeshers); err != nil {
		log.Warning("failed to cache smeshers: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	cache.LastUpdated.WithLabelValues(fmt.Sprintf("/smeshers?limit=%d&offset=%d", limit, offset)).SetToCurrentTime()

	return c.JSON(http.StatusOK, smeshers)
}

func SmeshersRefresh(c echo.Context) error {
	cc := c.(*ApiContext)

	go func() {
		smeshers, err := cc.StorageClient.GetSmeshers(cc.Storage, 1000, 0)
		if err != nil {
			log.Warning("failed to get smeshers: %v", err)
			return
		}

		for i := 0; i < len(smeshers.Smeshers); i += 20 {
			end := i + 20
			if end > len(smeshers.Smeshers) {
				end = len(smeshers.Smeshers)
			}
			if err = cc.Cache.Set(
				context.Background(),
				fmt.Sprintf("smeshers-%d-%d", 20, i),
				&storage.SmesherList{Smeshers: smeshers.Smeshers[i:end]},
			); err != nil {
				log.Warning("failed to cache smeshers: %v", err)
				return
			}
		}

		log.Info("smeshers refreshed")
		cache.LastUpdated.WithLabelValues("/refresh/smeshers").SetToCurrentTime()
	}()

	return c.NoContent(http.StatusOK)
}

func SmeshersByEpoch(c echo.Context) error {
	cc := c.(*ApiContext)

	epochId, err := strconv.Atoi(c.Param("epoch"))
	if err != nil || epochId < 0 {
		return c.NoContent(http.StatusBadRequest)
	}

	limit, offset := GetPagination(c)

	if cached, err := cc.Cache.Get(context.Background(),
		fmt.Sprintf("smeshers-epoch-%d-%d-%d", epochId, limit, offset), new(storage.SmesherList)); err == nil {
		return c.JSON(http.StatusOK, cached)
	}

	smeshers, err := cc.StorageClient.GetSmeshersByEpoch(cc.Storage, uint64(limit), uint64(offset), uint64(epochId))
	if err != nil {
		log.Warning("failed to get smeshers: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if err = cc.Cache.Set(context.Background(),
		fmt.Sprintf("smeshers-epoch-%d-%d-%d", epochId, limit, offset), smeshers); err != nil {
		log.Warning("failed to cache smeshers: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	cache.LastUpdated.WithLabelValues(
		fmt.Sprintf("/smeshers/%s?limit=%d&offset=%d", c.Param("epoch"), limit, offset)).SetToCurrentTime()

	return c.JSON(http.StatusOK, smeshers)
}

func SmeshersByEpochRefresh(c echo.Context) error {
	cc := c.(*ApiContext)

	epochId, err := strconv.Atoi(c.Param("epoch"))
	if err != nil || epochId < 0 {
		return c.NoContent(http.StatusBadRequest)
	}

	go func() {
		smeshers, err := cc.StorageClient.GetSmeshersByEpoch(cc.Storage, 1000, 0, uint64(epochId))
		if err != nil {
			log.Warning("failed to get smeshers: %v", err)
			return
		}

		for i := 0; i < len(smeshers.Smeshers); i += 20 {
			end := i + 20
			if end > len(smeshers.Smeshers) {
				end = len(smeshers.Smeshers)
			}
			if err = cc.Cache.Set(
				context.Background(),
				fmt.Sprintf("smeshers-%d-%d", 20, i),
				&storage.SmesherList{Smeshers: smeshers.Smeshers[i:end]},
			); err != nil {
				log.Warning("failed to cache smeshers: %v", err)
				return
			}
		}

		log.Info("smeshers by epoch %d refreshed", epochId)
		cache.LastUpdated.WithLabelValues("/refresh/smeshers/" + c.Param("epoch")).SetToCurrentTime()
	}()

	return c.NoContent(http.StatusOK)
}

func Smesher(c echo.Context) error {
	cc := c.(*ApiContext)

	smesherId := c.Param("smesherId")
	hash := types.HexToHash32(smesherId)

	if cached, err := cc.Cache.Get(context.Background(), "smesher-"+smesherId, new(*storage.Smesher)); err == nil {
		return c.JSON(http.StatusOK, cached)
	}

	smesher, err := cc.StorageClient.GetSmesher(cc.Storage, hash.Bytes())
	if err != nil {
		if err.Error() == "smesher not found" {
			return c.NoContent(http.StatusNotFound)
		}

		log.Warning("failed to get smesher: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if err = cc.Cache.Set(context.Background(), "smesher-"+smesherId, smesher,
		store.WithExpiration(cache.ShortExpiration)); err != nil {
		log.Warning("failed to cache smesher: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, smesher)
}
