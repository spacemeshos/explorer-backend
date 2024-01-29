package sql

import (
	"fmt"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/sql"
)

type DatabaseClient interface {
	GetLayer(db *sql.Database, lid types.LayerID, numLayers uint32) (*pb.Layer, error)
	GetLayerRewards(db *sql.Database, lid types.LayerID) (rst []*types.Reward, err error)
	GetAllRewards(db *sql.Database) (rst []*types.Reward, err error)
	AccountsSnapshot(db *sql.Database, lid types.LayerID) (rst []*types.Account, err error)
}

type Client struct{}

func Setup(path string) (db *sql.Database, err error) {
	db, err = sql.Open(fmt.Sprintf("file:%s?mode=ro", path),
		sql.WithConnections(16), sql.WithMigrations(nil))
	return
}
