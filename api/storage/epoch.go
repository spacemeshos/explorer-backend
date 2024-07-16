package storage

import (
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/sql"
	"github.com/spacemeshos/go-spacemesh/sql/atxs"
	"github.com/spacemeshos/go-spacemesh/sql/builder"
)

type EpochStats struct {
	TransactionsCount uint64 `json:"transactions_count"`
	ActivationsCount  uint64 `json:"activations_count"`
	RewardsCount      uint64 `json:"rewards_count"`
	RewardsSum        uint64 `json:"rewards_sum"`
	NumUnits          uint64 `json:"num_units"`
	SmeshersCount     uint64 `json:"smeshers_count"`
}

func (c *Client) GetEpochStats(db *sql.Database, epoch int64, layersPerEpoch int64) (*EpochStats, error) {
	stats := &EpochStats{
		TransactionsCount: 0,
		ActivationsCount:  0,
		RewardsCount:      0,
		RewardsSum:        0,
	}

	start := epoch * layersPerEpoch
	end := start + layersPerEpoch - 1

	_, err := db.Exec(`SELECT COUNT(*)
FROM (
  SELECT distinct id
  FROM transactions
  LEFT JOIN transactions_results_addresses
  ON transactions.id = transactions_results_addresses.tid
  WHERE layer >= ?1 and layer <= ?2
);`,
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, start)
			stmt.BindInt64(2, end)
		},
		func(stmt *sql.Statement) bool {
			stats.TransactionsCount = uint64(stmt.ColumnInt64(0))
			return true
		})
	if err != nil {
		return nil, err
	}

	ops := builder.Operations{
		Filter: []builder.Op{
			{
				Field: builder.Epoch,
				Token: builder.Eq,
				Value: epoch - 1,
			},
		},
	}
	count, err := atxs.CountAtxsByOps(db, ops)
	if err != nil {
		log.Err(err)
		return nil, err
	}
	stats.ActivationsCount = uint64(count)

	_, err = db.Exec(`SELECT COUNT(*), SUM(total_reward) FROM rewards WHERE layer >= ?1 and layer <= ?2`,
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, start)
			stmt.BindInt64(2, end)
		},
		func(stmt *sql.Statement) bool {
			stats.RewardsCount = uint64(stmt.ColumnInt64(0))
			stats.RewardsSum = uint64(stmt.ColumnInt64(1))
			return true
		})
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`SELECT SUM(effective_num_units) FROM (SELECT effective_num_units FROM atxs WHERE epoch = ?1)`,
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, epoch-1)
		},
		func(stmt *sql.Statement) bool {
			stats.NumUnits = uint64(stmt.ColumnInt64(0))
			return true
		})
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`SELECT COUNT(*) FROM (SELECT DISTINCT pubkey FROM atxs WHERE epoch = ?1)`,
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, epoch-1)
		},
		func(stmt *sql.Statement) bool {
			stats.SmeshersCount = uint64(stmt.ColumnInt64(0))
			return true
		})

	return stats, err
}
