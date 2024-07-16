package storage

import (
	"fmt"
	"github.com/spacemeshos/go-spacemesh/sql"
)

type DatabaseClient interface {
	GetLayerStats(db *sql.Database, lid int64) (*LayerStats, error)
	GetEpochStats(db *sql.Database, epoch int64, layersPerEpoch int64) (*EpochStats, error)
	GetSmeshers(db *sql.Database, limit, offset uint64) (*SmesherList, error)
	GetSmeshersByEpoch(db *sql.Database, limit, offset, epoch uint64) (*SmesherList, error)
	GetSmesher(db *sql.Database, pubkey []byte) (*Smesher, error)

	GetAccountsCount(db *sql.Database) (uint64, error)
	GetSmeshersCount(db *sql.Database) (uint64, error)
	GetLayersCount(db *sql.Database) (uint64, error)
	GetRewardsCount(db *sql.Database) (uint64, error)
	GetTransactionsCount(db *sql.Database) (uint64, error)
	GetTotalNumUnits(db *sql.Database) (uint64, error)
}

type Client struct{}

func Setup(path string) (db *sql.Database, err error) {
	db, err = sql.Open(fmt.Sprintf("file:%s?mode=ro", path),
		sql.WithConnections(16), sql.WithMigrations(nil))
	return
}
