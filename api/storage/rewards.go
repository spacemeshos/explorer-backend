package storage

import (
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/sql"
)

func (c *Client) GetRewardsSum(db *sql.Database) (sum uint64, count uint64, err error) {
	_, err = db.Exec(`SELECT COUNT(*), SUM(total_reward) FROM rewards`,
		func(stmt *sql.Statement) {
		},
		func(stmt *sql.Statement) bool {
			count = uint64(stmt.ColumnInt64(0))
			sum = uint64(stmt.ColumnInt64(1))
			return true
		})
	return
}

func (c *Client) GetRewardsSumByAddress(db *sql.Database, addr types.Address) (sum uint64, count uint64, err error) {
	_, err = db.Exec(`SELECT COUNT(*), SUM(total_reward) FROM rewards WHERE coinbase = ?1`,
		func(stmt *sql.Statement) {
			stmt.BindBytes(1, addr.Bytes())
		},
		func(stmt *sql.Statement) bool {
			count = uint64(stmt.ColumnInt64(0))
			sum = uint64(stmt.ColumnInt64(1))
			return true
		})
	return
}
