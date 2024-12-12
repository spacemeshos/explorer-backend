package storage

import "github.com/spacemeshos/go-spacemesh/sql"

func (c *Client) GetTotalNumUnits(db sql.Executor) (count uint64, err error) {
	_, err = db.Exec(`SELECT SUM(effective_num_units) FROM atxs;`,
		func(stmt *sql.Statement) {
		},
		func(stmt *sql.Statement) bool {
			count = uint64(stmt.ColumnInt64(0))
			return true
		})
	return
}
