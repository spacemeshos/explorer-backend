package storage

import (
	"fmt"

	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/sql"
	"github.com/spacemeshos/go-spacemesh/sql/statesql"
	"github.com/spacemeshos/go-spacemesh/timesync"
)

type DatabaseClient interface {
	Overview(db sql.Executor) (*Overview, error)

	GetLayerStats(db sql.Executor, lid int64) (*LayerStats, error)
	GetLayersCount(db sql.Executor) (uint64, error)

	GetEpochStats(db sql.Executor, epoch, layersPerEpoch int64) (*EpochStats, error)
	GetEpochDecentralRatio(db sql.Executor, epoch int64) (*EpochStats, error)

	GetSmeshers(db sql.Executor, limit, offset uint64) (*SmesherList, error)
	GetSmeshersByEpoch(db sql.Executor, limit, offset, epoch uint64) (*SmesherList, error)
	GetSmesher(db sql.Executor, pubkey []byte) (*Smesher, error)

	GetAccountsCount(db sql.Executor) (uint64, error)
	GetAccountsStats(db sql.Executor, addr types.Address) (*AccountStats, error)

	GetSmeshersCount(db sql.Executor) (uint64, error)
	GetSmeshersByEpochCount(db sql.Executor, epoch uint64) (uint64, error)

	GetRewardsSum(db sql.Executor) (uint64, uint64, error)
	GetRewardsSumByAddress(db sql.Executor, addr types.Address) (sum, count uint64, err error)

	GetTransactionsCount(db sql.Executor) (uint64, error)
	GetTotalNumUnits(db sql.Executor) (uint64, error)

	GetCirculation(db sql.Executor) (*Circulation, error)
}

type Client struct {
	NodeClock     *timesync.NodeClock
	Testnet       bool
	LabelsPerUnit uint64
	BitsPerLabel  uint64
}

func Setup(path string) (db sql.StateDatabase, err error) {
	return statesql.Open(fmt.Sprintf("file:%s?mode=ro", path), sql.WithConnections(16))
}
