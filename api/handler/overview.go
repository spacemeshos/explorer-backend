package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/api/cache"
	"github.com/spacemeshos/go-spacemesh/log"
	"net/http"
)

type OverviewResp struct {
	AccountsCount     uint64 `json:"accounts_count"`
	SmeshersCount     uint64 `json:"smeshers_count"`
	LayersCount       uint64 `json:"layers_count"`
	RewardsCount      uint64 `json:"rewards_count"`
	RewardsSum        uint64 `json:"rewards_sum"`
	TransactionsCount uint64 `json:"transactions_count"`
	NumUnits          uint64 `json:"num_units"`
}

func Overview(c echo.Context) error {
	cc := c.(*ApiContext)

	if cached, ok := cache.Cache.Get("overview"); ok {
		return c.JSON(http.StatusOK, cached)
	}

	overviewResp := OverviewResp{}
	accountsCount, err := cc.StorageClient.GetAccountsCount(cc.Storage)
	if err != nil {
		log.Warning("failed to get accounts count: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	overviewResp.AccountsCount = accountsCount

	smeshersCount, err := cc.StorageClient.GetSmeshersCount(cc.Storage)
	if err != nil {
		log.Warning("failed to get smeshers count: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	overviewResp.SmeshersCount = smeshersCount

	layersCount, err := cc.StorageClient.GetLayersCount(cc.Storage)
	if err != nil {
		log.Warning("failed to get layers count: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	overviewResp.LayersCount = layersCount

	rewardsSum, rewardsCount, err := cc.StorageClient.GetRewardsSum(cc.Storage)
	if err != nil {
		log.Warning("failed to get rewards count: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	overviewResp.RewardsSum = rewardsSum
	overviewResp.RewardsCount = rewardsCount

	transactionsCount, err := cc.StorageClient.GetTransactionsCount(cc.Storage)
	if err != nil {
		log.Warning("failed to get transactions count: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	overviewResp.TransactionsCount = transactionsCount

	numUnits, err := cc.StorageClient.GetTotalNumUnits(cc.Storage)
	if err != nil {
		log.Warning("failed to get num units count: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	overviewResp.NumUnits = numUnits

	cache.Cache.Set("overview", overviewResp, 0)
	return c.JSON(http.StatusOK, overviewResp)
}
