package sql

import (
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/sql"
	"github.com/spacemeshos/go-spacemesh/sql/accounts"
)

func (c *Client) AccountsSnapshot(db *sql.Database, lid types.LayerID) (rst []*types.Account, err error) {
	return accounts.Snapshot(db, lid)
}
