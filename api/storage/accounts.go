package storage

import "github.com/spacemeshos/go-spacemesh/sql"

func (c *Client) GetAccountsCount(db *sql.Database) (uint64, error) {
	var total uint64
	_, err := db.Exec(`SELECT COUNT(DISTINCT address) FROM accounts`,
		func(stmt *sql.Statement) {
		},
		func(stmt *sql.Statement) bool {
			total = uint64(stmt.ColumnInt64(0))
			return true
		})
	return total, err
}
