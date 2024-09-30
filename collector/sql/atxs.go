package sql

import (
	sqlite "github.com/go-llsqlite/crawshaw"
	"github.com/spacemeshos/explorer-backend/utils"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/sql"
	"github.com/spacemeshos/go-spacemesh/sql/atxs"
	"time"
)

// Query to retrieve ATXs.
// Can't use inner join for the ATX blob here b/c this will break
// filters that refer to the id column.
const fieldsQuery = `select
atxs.id, atxs.nonce, atxs.base_tick_height, atxs.tick_count, atxs.pubkey, atxs.effective_num_units,
atxs.received, atxs.epoch, atxs.sequence, atxs.coinbase, atxs.validity, atxs.commitment_atx, atxs.weight,
atxs.marriage_atx`

const fullQuery = fieldsQuery + ` from atxs`

type decoderCallback func(*types.ActivationTx) bool

func decoder(fn decoderCallback) sql.Decoder {
	return func(stmt *sql.Statement) bool {
		var (
			a  types.ActivationTx
			id types.ATXID
		)
		stmt.ColumnBytes(0, id[:])
		a.SetID(id)
		a.VRFNonce = types.VRFPostIndex(stmt.ColumnInt64(1))
		a.BaseTickHeight = uint64(stmt.ColumnInt64(2))
		a.TickCount = uint64(stmt.ColumnInt64(3))
		stmt.ColumnBytes(4, a.SmesherID[:])
		a.NumUnits = uint32(stmt.ColumnInt32(5))
		// Note: received is assigned `0` for checkpointed ATXs.
		// We treat `0` as 'zero time'.
		// We could use `NULL` instead, but the column has "NOT NULL" constraint.
		// In future, consider changing the schema to allow `NULL` for received.
		if received := stmt.ColumnInt64(6); received == 0 {
			a.SetGolden()
		} else {
			a.SetReceived(time.Unix(0, received).Local())
		}
		a.PublishEpoch = types.EpochID(uint32(stmt.ColumnInt(7)))
		a.Sequence = uint64(stmt.ColumnInt64(8))
		stmt.ColumnBytes(9, a.Coinbase[:])
		a.SetValidity(types.Validity(stmt.ColumnInt(10)))
		if stmt.ColumnType(11) != sqlite.SQLITE_NULL {
			a.CommitmentATX = new(types.ATXID)
			stmt.ColumnBytes(11, a.CommitmentATX[:])
		}
		a.Weight = uint64(stmt.ColumnInt64(12))
		if stmt.ColumnType(13) != sqlite.SQLITE_NULL {
			a.MarriageATX = new(types.ATXID)
			stmt.ColumnBytes(13, a.MarriageATX[:])
		}

		return fn(&a)
	}
}

func (c *Client) GetAtxsReceivedAfter(db sql.Executor, ts int64, fn func(tx *types.ActivationTx) bool) error {
	var derr error
	_, err := db.Exec(
		fullQuery+` WHERE received > ?1`,
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, ts)
		},
		decoder(func(atx *types.ActivationTx) bool {
			if atx != nil {
				return fn(atx)
			}
			return true
		}),
	)
	if err != nil {
		return err
	}
	return derr
}

func (c *Client) GetAtxsByEpoch(db sql.Executor, epoch int64, fn func(tx *types.ActivationTx) bool) error {
	var derr error
	_, err := db.Exec(
		fullQuery+` WHERE epoch = ?1 ORDER BY epoch asc, id asc`,
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, epoch)
		},
		decoder(func(atx *types.ActivationTx) bool {
			if atx != nil {
				return fn(atx)
			}
			return true
		}),
	)
	if err != nil {
		return err
	}
	return derr
}

func (c *Client) CountAtxsByEpoch(db sql.Executor, epoch int64) (int, error) {
	var totalCount int
	_, err := db.Exec(
		`SELECT COUNT(*) FROM atxs WHERE epoch = ?1`,
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, epoch)
		}, func(stmt *sql.Statement) bool {
			totalCount = stmt.ColumnInt(0)
			return true
		})
	if err != nil {
		return 0, err
	}
	return totalCount, nil
}

func (c *Client) GetAtxsByEpochPaginated(db sql.Executor, epoch, limit, offset int64, fn func(tx *types.ActivationTx) bool) error {
	var derr error
	_, err := db.Exec(
		fullQuery+` WHERE epoch = ?1 ORDER BY epoch asc, id asc LIMIT ?2 OFFSET ?3`,
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, epoch)
			stmt.BindInt64(2, limit)
			stmt.BindInt64(3, offset)
		},
		decoder(func(atx *types.ActivationTx) bool {
			if atx != nil {
				return fn(atx)
			}
			return true
		}),
	)
	if err != nil {
		return err
	}
	return derr
}

func (c *Client) GetAtxById(db sql.Executor, id string) (*types.ActivationTx, error) {
	idBytes, err := utils.StringToBytes(id)
	if err != nil {
		return nil, err
	}

	var atxId types.ATXID
	copy(atxId[:], idBytes)

	atx, err := atxs.Get(db, atxId)
	if err != nil {
		return nil, err
	}

	return atx, nil
}
