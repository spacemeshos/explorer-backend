package storage

import (
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/sql"
)

type SmesherList struct {
	Smeshers []Smesher `json:"smeshers"`
}

type Smesher struct {
	Pubkey       types.NodeID `json:"pubkey"`
	Coinbase     string       `json:"coinbase,omitempty"`
	NumUnits     uint64       `json:"num_units,omitempty"`
	Atxs         uint64       `json:"atxs"`
	RewardsCount uint64       `json:"rewards_count,omitempty"`
	RewardsSum   uint64       `json:"rewards_sum,omitempty"`
}

func (c *Client) GetSmeshers(db *sql.Database, limit, offset uint64) (*SmesherList, error) {
	smesherList := &SmesherList{
		Smeshers: []Smesher{},
	}

	_, err := db.Exec(`SELECT pubkey, COUNT(*) as atxs FROM atxs GROUP BY pubkey ORDER BY pubkey ASC, epoch DESC LIMIT ?1 OFFSET ?2;`,
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, int64(limit))
			stmt.BindInt64(2, int64(offset))
		},
		func(stmt *sql.Statement) bool {
			var smesher Smesher
			stmt.ColumnBytes(0, smesher.Pubkey[:])
			smesher.Atxs = uint64(stmt.ColumnInt64(1))
			smesherList.Smeshers = append(smesherList.Smeshers, smesher)
			return true
		})
	if err != nil {
		return nil, err
	}

	return smesherList, err
}

func (c *Client) GetSmeshersByEpoch(db *sql.Database, limit, offset, epoch uint64) (*SmesherList, error) {
	smesherList := &SmesherList{
		Smeshers: []Smesher{},
	}

	_, err := db.Exec(`SELECT DISTINCT pubkey, COUNT(*) as atxs FROM atxs WHERE epoch = ?1 GROUP BY pubkey ORDER BY pubkey ASC, epoch DESC LIMIT ?2 OFFSET ?3;`,
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, int64(epoch-1))
			stmt.BindInt64(2, int64(limit))
			stmt.BindInt64(3, int64(offset))
		},
		func(stmt *sql.Statement) bool {
			var smesher Smesher
			stmt.ColumnBytes(0, smesher.Pubkey[:])
			smesher.Atxs = uint64(stmt.ColumnInt64(1))
			smesherList.Smeshers = append(smesherList.Smeshers, smesher)
			return true
		})
	if err != nil {
		return nil, err
	}

	return smesherList, err
}

func (c *Client) GetSmeshersCount(db *sql.Database) (count uint64, err error) {
	_, err = db.Exec(`SELECT COUNT(*) FROM (SELECT DISTINCT pubkey FROM atxs)`,
		func(stmt *sql.Statement) {
		},
		func(stmt *sql.Statement) bool {
			count = uint64(stmt.ColumnInt64(0))
			return true
		})
	return
}

func (c *Client) GetSmesher(db *sql.Database, pubkey []byte) (smesher *Smesher, err error) {
	smesher = &Smesher{}
	_, err = db.Exec(`SELECT pubkey, coinbase, effective_num_units, COUNT(*) as atxs FROM atxs WHERE pubkey = ?1 GROUP BY pubkey ORDER BY epoch DESC LIMIT 1;`,
		func(stmt *sql.Statement) {
			stmt.BindBytes(1, pubkey)
		},
		func(stmt *sql.Statement) bool {
			stmt.ColumnBytes(0, smesher.Pubkey[:])
			var coinbase types.Address
			stmt.ColumnBytes(1, coinbase[:])
			smesher.Coinbase = coinbase.String()
			smesher.NumUnits = uint64(stmt.ColumnInt64(2))
			smesher.Atxs = uint64(stmt.ColumnInt64(3))
			return true
		})
	if err != nil {
		return
	}

	_, err = db.Exec(`SELECT COUNT(*), SUM(total_reward) FROM rewards WHERE pubkey=?1`,
		func(stmt *sql.Statement) {
			stmt.BindBytes(1, pubkey)
		},
		func(stmt *sql.Statement) bool {
			smesher.RewardsCount = uint64(stmt.ColumnInt64(0))
			smesher.RewardsSum = uint64(stmt.ColumnInt64(1))
			return true
		})
	return
}
