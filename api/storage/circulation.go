package storage

import (
	"github.com/spacemeshos/economics/vesting"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/sql"
)

type Circulation struct {
	Circulation uint64 `json:"circulation"`
}

func (c *Client) GetCirculation(db sql.Executor) (*Circulation, error) {
	circulation := &Circulation{
		Circulation: 0,
	}
	if !c.Testnet {
		accumulatedVest := vesting.AccumulatedVestAtLayer(c.NodeClock.CurrentLayer().Uint32())
		circulation.Circulation = accumulatedVest
	}

	rewardsSum, _, err := c.GetRewardsSum(db)
	if err != nil {
		log.Warning("failed to get rewards count: %v", err)
		return nil, err
	}
	circulation.Circulation += rewardsSum

	return circulation, nil
}
