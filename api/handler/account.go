package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/log"
	"net/http"
)

func AccountStats(c echo.Context) error {
	cc := c.(*ApiContext)

	address := c.Param("address")
	addr, err := types.StringToAddress(address)
	if err != nil {
		log.Warning("failed to parse account address: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	accountStats, err := cc.StorageClient.GetAccountsStats(cc.Storage, addr)
	if err != nil {
		log.Warning("failed to get account stats: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, accountStats)
}
