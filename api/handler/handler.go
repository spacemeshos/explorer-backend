package handler

import (
	"github.com/eko/gocache/lib/v4/marshaler"
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/api/storage"
	"github.com/spacemeshos/go-spacemesh/sql"
	"strconv"
)

type ApiContext struct {
	echo.Context
	Storage        *sql.Database
	StorageClient  storage.DatabaseClient
	LayersPerEpoch int64
	Cache          *marshaler.Marshaler
}

func GetPagination(c echo.Context) (limit, offset int64) {
	limit = 20
	offset = 0
	if page := c.QueryParam("limit"); page != "" {
		limit, _ = strconv.ParseInt(page, 10, 32)
		if limit <= 0 {
			limit = 0
		}
	}
	if size := c.QueryParam("offset"); size != "" {
		offset, _ = strconv.ParseInt(size, 10, 32)
		if offset <= 0 {
			offset = 0
		}
	}
	return limit, offset
}
