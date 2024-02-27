package sql

import (
	"fmt"
	"github.com/spacemeshos/go-spacemesh/codec"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/sql"
	"time"
)

const fullQuery = `select id, atx, base_tick_height, tick_count, pubkey,
	effective_num_units, received, epoch, sequence, coinbase, validity from atxs`

type decoderCallback func(*types.VerifiedActivationTx, error) bool

func decoder(fn decoderCallback) sql.Decoder {
	return func(stmt *sql.Statement) bool {
		var (
			a  types.ActivationTx
			id types.ATXID
		)
		stmt.ColumnBytes(0, id[:])
		checkpointed := stmt.ColumnLen(1) == 0
		if !checkpointed {
			if _, err := codec.DecodeFrom(stmt.ColumnReader(1), &a); err != nil {
				return fn(nil, fmt.Errorf("decode %w", err))
			}
		}
		a.SetID(id)
		baseTickHeight := uint64(stmt.ColumnInt64(2))
		tickCount := uint64(stmt.ColumnInt64(3))
		stmt.ColumnBytes(4, a.SmesherID[:])
		effectiveNumUnits := uint32(stmt.ColumnInt32(5))
		a.SetEffectiveNumUnits(effectiveNumUnits)
		if checkpointed {
			a.SetGolden()
			a.NumUnits = effectiveNumUnits
			a.SetReceived(time.Time{})
		} else {
			a.SetReceived(time.Unix(0, stmt.ColumnInt64(6)).Local())
		}
		a.PublishEpoch = types.EpochID(uint32(stmt.ColumnInt(7)))
		a.Sequence = uint64(stmt.ColumnInt64(8))
		stmt.ColumnBytes(9, a.Coinbase[:])
		a.SetValidity(types.Validity(stmt.ColumnInt(10)))
		v, err := a.Verify(baseTickHeight, tickCount)
		if err != nil {
			return fn(nil, err)
		}
		return fn(v, nil)
	}
}

func (c *Client) GetAtxsReceivedAfter(db *sql.Database, ts int64, fn func(tx *types.VerifiedActivationTx) bool) error {
	var derr error
	_, err := db.Exec(
		fullQuery+` WHERE received > ?1 ORDER BY epoch asc, id asc`,
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, ts)
		},
		decoder(func(atx *types.VerifiedActivationTx, err error) bool {
			if atx != nil {
				return fn(atx)
			}
			derr = err
			return derr == nil
		}),
	)
	if err != nil {
		return err
	}
	return derr
}
