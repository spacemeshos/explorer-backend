package collector

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/log"
)

func (c *Collector) GetAccountState(address string) (uint64, uint64, error) {
	req := &pb.AccountRequest{AccountId: &pb.AccountId{Address: address}}

	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.globalClient.Account(ctx, req)
	if err != nil {
		log.Err(fmt.Errorf("cannot get account info: %v", err))
		return 0, 0, err
	}

	if res.AccountWrapper == nil || res.AccountWrapper.StateCurrent == nil || res.AccountWrapper.StateCurrent.Balance == nil {
		return 0, 0, errors.New("Bad result")
	}

	return res.AccountWrapper.StateCurrent.Balance.Value, res.AccountWrapper.StateCurrent.Counter, nil
}
