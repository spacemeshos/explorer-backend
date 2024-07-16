package storage

import "github.com/spacemeshos/go-spacemesh/sql"

func (c *Client) GetRewardsCount(db *sql.Database) (count uint64, err error) {
	_, err = db.Exec(`SELECT COUNT(*) FROM rewards`,
		func(stmt *sql.Statement) {
		},
		func(stmt *sql.Statement) bool {
			count = uint64(stmt.ColumnInt64(0))
			return true
		})
	return
}
