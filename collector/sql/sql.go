package sql

import (
	"fmt"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/sql"
	"github.com/spacemeshos/go-spacemesh/sql/statesql"
)

type DatabaseClient interface {
	GetLayer(db sql.Executor, lid types.LayerID, numLayers uint32) (*pb.Layer, error)
	GetLayerRewards(db sql.Executor, lid types.LayerID) (rst []*types.Reward, err error)
	GetAllRewards(db sql.Executor) (rst []*types.Reward, err error)
	AccountsSnapshot(db sql.Executor, lid types.LayerID) (rst []*types.Account, err error)
	GetAtxsReceivedAfter(db sql.Executor, ts int64, fn func(tx *types.ActivationTx) bool) error
	GetAtxsByEpoch(db sql.Executor, epoch int64, fn func(tx *types.ActivationTx) bool) error
	CountAtxsByEpoch(db sql.Executor, epoch int64) (int, error)
	GetAtxsByEpochPaginated(db sql.Executor, epoch, limit, offset int64, fn func(tx *types.ActivationTx) bool) error
	GetAtxById(db sql.Executor, id string) (*types.ActivationTx, error)
}

type Client struct{}

func Setup(path string) (db sql.StateDatabase, err error) {
	db, err = statesql.Open(fmt.Sprintf("file:%s?mode=ro", path),
		sql.WithConnections(16), sql.WithMigrationsDisabled())
	return
}
