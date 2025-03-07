package handler

import (
	"strconv"

	"github.com/eko/gocache/lib/v4/marshaler"
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/go-spacemesh/sql"

	"github.com/spacemeshos/explorer-backend/api/storage"
)

type ApiContext struct {
	echo.Context
	Storage        sql.StateDatabase
	StorageClient  storage.DatabaseClient
	LayersPerEpoch int64
	Cache          *marshaler.Marshaler
}

func GetPagination(c echo.Context) (limit, offset int64) {
	limit = 20
	offset = 0
	if size := c.QueryParam("offset"); size != "" {
		offset, _ = strconv.ParseInt(size, 10, 32)
		if offset <= 0 {
			offset = 0
		}
	}
	return limit, offset
}
