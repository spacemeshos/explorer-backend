package storage

import (
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/sql"
)

type Overview struct {
	AccountsCount     uint64 `json:"accounts_count"`
	SmeshersCount     uint64 `json:"smeshers_count"`
	LayersCount       uint64 `json:"layers_count"`
	RewardsCount      uint64 `json:"rewards_count"`
	RewardsSum        uint64 `json:"rewards_sum"`
	TransactionsCount uint64 `json:"transactions_count"`
	NumUnits          uint64 `json:"num_units"`
}

func (c *Client) Overview(db sql.Executor) (*Overview, error) {
	overview := &Overview{}
	accountsCount, err := c.GetAccountsCount(db)
	if err != nil {
		log.Warning("failed to get accounts count: %v", err)
		return nil, err
	}
	overview.AccountsCount = accountsCount

	smeshersCount, err := c.GetSmeshersCount(db)
	if err != nil {
		log.Warning("failed to get smeshers count: %v", err)
		return nil, err
	}
	overview.SmeshersCount = smeshersCount

	layersCount, err := c.GetLayersCount(db)
	if err != nil {
		log.Warning("failed to get layers count: %v", err)
		return nil, err
	}
	overview.LayersCount = layersCount

	rewardsSum, rewardsCount, err := c.GetRewardsSum(db)
	if err != nil {
		log.Warning("failed to get rewards count: %v", err)
		return nil, err
	}
	overview.RewardsSum = rewardsSum
	overview.RewardsCount = rewardsCount

	transactionsCount, err := c.GetTransactionsCount(db)
	if err != nil {
		log.Warning("failed to get transactions count: %v", err)
		return nil, err
	}
	overview.TransactionsCount = transactionsCount

	numUnits, err := c.GetTotalNumUnits(db)
	if err != nil {
		log.Warning("failed to get num units count: %v", err)
		return nil, err
	}
	overview.NumUnits = numUnits

	return overview, nil
}
