package storage

import (
	"fmt"

	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/sql"
	"github.com/spacemeshos/go-spacemesh/timesync"
)

type DatabaseClient interface {
	Overview(db *sql.Database) (*Overview, error)

	GetLayerStats(db *sql.Database, lid int64) (*LayerStats, error)
	GetLayersCount(db *sql.Database) (uint64, error)

	GetEpochStats(db *sql.Database, epoch, layersPerEpoch int64) (*EpochStats, error)
	GetEpochDecentralRatio(db *sql.Database, epoch int64) (*EpochStats, error)

	GetSmeshers(db *sql.Database, limit, offset uint64) (*SmesherList, error)
	GetSmeshersByEpoch(db *sql.Database, limit, offset, epoch uint64) (*SmesherList, error)
	GetSmesher(db *sql.Database, pubkey []byte) (*Smesher, error)

	GetAccountsCount(db *sql.Database) (uint64, error)
	GetAccountsStats(db *sql.Database, addr types.Address) (*AccountStats, error)

	GetSmeshersCount(db *sql.Database) (uint64, error)
	GetSmeshersByEpochCount(db *sql.Database, epoch uint64) (uint64, error)

	GetRewardsSum(db *sql.Database) (uint64, uint64, error)
	GetRewardsSumByAddress(db *sql.Database, addr types.Address) (sum, count uint64, err error)

	GetTransactionsCount(db *sql.Database) (uint64, error)
	GetTotalNumUnits(db *sql.Database) (uint64, error)

	GetCirculation(db *sql.Database) (*Circulation, error)
}

type Client struct {
	NodeClock     *timesync.NodeClock
	Testnet       bool
	LabelsPerUnit uint64
	BitsPerLabel  uint64
}

func Setup(path string) (db *sql.Database, err error) {
	db, err = sql.Open(fmt.Sprintf("file:%s?mode=ro", path),
		sql.WithConnections(16), sql.WithMigrations(nil))
	return
}
