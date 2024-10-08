package sql

import (
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/sql"
)

func (c *Client) GetLayerRewards(db sql.Executor, lid types.LayerID) (rst []*types.Reward, err error) {
	_, err = db.Exec("select coinbase, layer, total_reward, layer_reward, pubkey from rewards where layer = ?1;",
		func(stmt *sql.Statement) {
			stmt.BindInt64(1, int64(lid))
		}, func(stmt *sql.Statement) bool {
			addrBytes := stmt.ColumnViewBytes(0)

			var addr types.Address
			copy(addr[:], addrBytes)

			reward := &types.Reward{
				Coinbase:    addr,
				Layer:       types.LayerID(uint32(stmt.ColumnInt64(1))),
				TotalReward: uint64(stmt.ColumnInt64(2)),
				LayerReward: uint64(stmt.ColumnInt64(3)),
				SmesherID:   types.BytesToNodeID(stmt.ColumnViewBytes(4)),
			}
			rst = append(rst, reward)
			return true
		})
	return
}

func (c *Client) GetAllRewards(db sql.Executor) (rst []*types.Reward, err error) {
	_, err = db.Exec("select coinbase, layer, total_reward, layer_reward, pubkey from rewards;",
		nil, func(stmt *sql.Statement) bool {
			addrBytes := stmt.ColumnViewBytes(0)

			var addr types.Address
			copy(addr[:], addrBytes)

			reward := &types.Reward{
				Coinbase:    addr,
				Layer:       types.LayerID(uint32(stmt.ColumnInt64(1))),
				TotalReward: uint64(stmt.ColumnInt64(2)),
				LayerReward: uint64(stmt.ColumnInt64(3)),
				SmesherID:   types.BytesToNodeID(stmt.ColumnViewBytes(4)),
			}
			rst = append(rst, reward)
			return true
		})
	return
}
