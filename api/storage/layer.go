package storage

import (
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/sql"
	"github.com/spacemeshos/go-spacemesh/sql/builder"
	"github.com/spacemeshos/go-spacemesh/sql/transactions"
)

type LayerStats struct {
	TransactionsCount uint64 `json:"transactions_count"`
	TransactionsSum   uint64 `json:"transactions_sum"`
	RewardsCount      uint64 `json:"rewards_count"`
	RewardsSum        uint64 `json:"rewards_sum"`
}

func (c *Client) GetLayerStats(db *sql.Database, lid int64) (*LayerStats, error) {
	stats := &LayerStats{
		TransactionsCount: 0,
		TransactionsSum:   0,
		RewardsCount:      0,
		RewardsSum:        0,
	}
	ops := builder.Operations{
		Filter: []builder.Op{
			{
				Field: builder.Layer,
				Token: builder.Eq,
				Value: lid,
			},
		},
	}
	err := transactions.IterateTransactionsOps(db, ops, func(tx *types.MeshTransaction,
		result *types.TransactionResult,
	) bool {
		contents, _, err := toTxContents(tx.Raw)
		if err != nil {
			return false
		}

		if contents.GetSend() != nil {
			stats.TransactionsSum += contents.GetSend().GetAmount()
		}

		stats.TransactionsCount++
		return true
	})
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`SELECT COUNT(*), SUM(total_reward) FROM rewards WHERE layer=?1`,
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, lid)
		},
		func(stmt *sql.Statement) bool {
			stats.RewardsCount = uint64(stmt.ColumnInt64(0))
			stats.RewardsSum = uint64(stmt.ColumnInt64(1))
			return true
		})
	if err != nil {
		return nil, err
	}

	return stats, err
}

func (c *Client) GetLayersCount(db *sql.Database) (count uint64, err error) {
	_, err = db.Exec(`SELECT COUNT(*) FROM layers`,
		func(stmt *sql.Statement) {
		},
		func(stmt *sql.Statement) bool {
			count = uint64(stmt.ColumnInt64(0))
			return true
		})
	return
}
