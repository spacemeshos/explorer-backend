package collector

import (
	"context"
	"fmt"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/log"
	"io"
)

func (c *Collector) transactionsPump() error {
	epochNumLayers := c.listener.GetEpochNumLayers()
	lastLayer := c.listener.GetLastLayer(context.Background())
	currentEpoch := lastLayer / epochNumLayers

	req := pb.TransactionResultsRequest{
		Start: (currentEpoch - 2) * epochNumLayers,
		Watch: true,
	}

	log.Info("Start transactions pump")
	defer func() {
		c.notify <- -streamType_transactions
		log.Info("Stop transactions pump")
	}()

	c.notify <- +streamType_transactions

	stream, err := c.transactionsClient.StreamResults(context.Background(), &req)
	if err != nil {
		log.Err(fmt.Errorf("cannot get transactions stream results: %v", err))
		return err
	}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			return err
		}
		if err != nil {
			log.Err(fmt.Errorf("cannot receive transaction result: %v", err))
			return err
		}
		if response == nil {
			continue
		}

		state, err := c.transactionsClient.TransactionsState(context.TODO(), &pb.TransactionsStateRequest{
			TransactionId:       []*pb.TransactionId{{Id: response.Tx.Id}},
			IncludeTransactions: false,
		})
		if err != nil {
			log.Err(fmt.Errorf("cannot receive transaction state: %v", err))
			return err
		}

		if len(state.GetTransactionsState()) > 0 {
			c.listener.OnTransactionResult(response, state.GetTransactionsState()[0])
		}
	}
}
