package handler

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/api/storage"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/log"
	"net/http"
	"strconv"
)

func Smeshers(c echo.Context) error {
	cc := c.(*ApiContext)
	limit, offset := GetPagination(c)

	smeshers, err := cc.StorageClient.GetSmeshers(cc.Storage, uint64(limit), uint64(offset))
	if err != nil {
		log.Warning("failed to get smeshers: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, smeshers)
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

	cc.Cache.Set(context.Background(), fmt.Sprintf("smeshers-epoch-%d-%d-%d", epochId, limit, offset), smeshers)
	return c.JSON(http.StatusOK, smeshers)
}

func Smesher(c echo.Context) error {
	cc := c.(*ApiContext)

	smesherId := c.Param("smesherId")
	hash := types.HexToHash32(smesherId)

	smesher, err := cc.StorageClient.GetSmesher(cc.Storage, hash.Bytes())
	if err != nil {
		if err.Error() == "smesher not found" {
			return c.NoContent(http.StatusNotFound)
		}

		log.Warning("failed to get smesher: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, smesher)
}
