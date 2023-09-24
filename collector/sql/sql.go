package sql

import (
	"fmt"
	"github.com/spacemeshos/go-spacemesh/sql"
)

func Setup(path string) (db *sql.Database, err error) {
	db, err = sql.Open(fmt.Sprintf("file:%s?mode=ro", path),
		sql.WithConnections(16), sql.WithMigrations(nil))
	return
}
