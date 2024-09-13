package storage

import (
	"math"

	"github.com/spacemeshos/economics/constants"
	"github.com/spacemeshos/explorer-backend/utils"

	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/sql"
	"github.com/spacemeshos/go-spacemesh/sql/atxs"
	"github.com/spacemeshos/go-spacemesh/sql/builder"
)

type EpochStats struct {
	TransactionsCount uint64 `json:"transactions_count,omitempty"`
	ActivationsCount  uint64 `json:"activations_count,omitempty"`
	RewardsCount      uint64 `json:"rewards_count,omitempty"`
	RewardsSum        uint64 `json:"rewards_sum,omitempty"`
	NumUnits          uint64 `json:"num_units,omitempty"`
	SmeshersCount     uint64 `json:"smeshers_count,omitempty"`
	Decentral         uint64 `json:"decentral,omitempty"`
	VestedAmount      uint64 `json:"vested_amount,omitempty"`
	AccountsCount     uint64 `json:"accounts_count,omitempty"`
}

func (c *Client) GetEpochStats(db *sql.Database, epoch, layersPerEpoch int64) (*EpochStats, error) {
	stats := &EpochStats{
		TransactionsCount: 0,
		ActivationsCount:  0,
		RewardsCount:      0,
		RewardsSum:        0,
	}

	start := epoch * layersPerEpoch
	end := start + layersPerEpoch - 1
	currentEpoch := c.NodeClock.CurrentLayer().Uint32() / uint32(layersPerEpoch)

	if !c.Testnet && end >= constants.VestStart {
		vestStartEpoch := constants.VestStart / layersPerEpoch
		if epoch == int64(currentEpoch) {
			stats.VestedAmount = (uint64(c.NodeClock.CurrentLayer().Uint32()) -
				uint64(start-1)) * constants.VestPerLayer
		} else if epoch == vestStartEpoch {
			stats.VestedAmount = uint64(end-constants.VestStart) * constants.VestPerLayer
		} else {
			stats.VestedAmount = uint64(layersPerEpoch) * constants.VestPerLayer
		}
	}

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
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`SELECT COUNT(DISTINCT address)
								FROM transactions_results_addresses
								WHERE tid IN (
									SELECT id FROM transactions WHERE layer >= ?1 AND layer <= ?2)`,
		func(statement *sql.Statement) {
			statement.BindInt64(1, start)
			statement.BindInt64(2, end)
		},
		func(statement *sql.Statement) bool {
			stats.AccountsCount = uint64(statement.ColumnInt64(0))
			return true
		})

	return stats, err
}

func (c *Client) GetEpochDecentralRatio(db *sql.Database, epoch int64) (*EpochStats, error) {
	stats := &EpochStats{
		Decentral: 0,
	}

	_, err := db.Exec(`SELECT COUNT(*) FROM (SELECT DISTINCT pubkey FROM atxs WHERE epoch = ?1)`,
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, epoch-1)
		},
		func(stmt *sql.Statement) bool {
			stats.SmeshersCount = uint64(stmt.ColumnInt64(0))
			return true
		})
	if err != nil {
		return nil, err
	}

	a := math.Min(float64(stats.SmeshersCount), 1e4)
	// pubkey: commitment size
	smeshers := make(map[string]uint64)
	_, err = db.Exec(`SELECT pubkey, effective_num_units FROM atxs WHERE epoch = ?1`,
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, epoch-1)
		},
		func(stmt *sql.Statement) bool {
			var smesher types.NodeID
			stmt.ColumnBytes(0, smesher[:])
			smeshers[smesher.String()] = uint64(stmt.ColumnInt64(1)) * ((c.BitsPerLabel * c.LabelsPerUnit) / 8)
			return true
		})
	if err != nil {
		return nil, err
	}

	stats.Decentral = uint64(100.0 * (0.5*(a*a)/1e8 + 0.5*(1.0-utils.Gini(smeshers))))

	return stats, nil
}
