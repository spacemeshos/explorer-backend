package storage

import (
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/sql"
	"github.com/spacemeshos/go-spacemesh/sql/builder"
	"github.com/spacemeshos/go-spacemesh/sql/transactions"
)

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

type AccountStats struct {
	Account  string `json:"account"`
	Received uint64 `json:"received"`
	Sent     uint64 `json:"sent"`
}

func (c *Client) GetAccountsStats(db *sql.Database, addr types.Address) (*AccountStats, error) {
	stats := &AccountStats{
		Account:  addr.String(),
		Received: 0,
		Sent:     0,
	}

	ops := builder.Operations{
		Filter: []builder.Op{
			{
				Group: []builder.Op{
					{
						Field: builder.Address,
						Token: builder.Eq,
						Value: addr.Bytes(),
					},
					{
						Field: builder.Principal,
						Token: builder.Eq,
						Value: addr.Bytes(),
					},
				},
				GroupOperator: builder.Or,
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
			if contents.GetSend().GetDestination() == addr.String() {
				stats.Received += contents.GetSend().GetAmount()
			} else {
				stats.Sent += contents.GetSend().GetAmount()
			}
		}
		return true
	})
	if err != nil {
		return nil, err
	}

	return stats, nil
}
