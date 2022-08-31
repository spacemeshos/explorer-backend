package collector

import (
    "context"
    "errors"
    "io"
    "time"

    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/go-spacemesh/log"
    "github.com/spacemeshos/explorer-backend/utils"
)

func (c *Collector) GetAccountState(address string) (uint64, uint64, error) {
	req := &pb.AccountRequest{AccountId: &pb.AccountId{Address: address}}

    // set timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    res, err := c.globalClient.Account(ctx, req)
    if err != nil {
        log.Error("cannot get account info: %v", err)
        return 0, 0, err
    }

    if res.AccountWrapper == nil || res.AccountWrapper.StateCurrent == nil || res.AccountWrapper.StateCurrent.Balance == nil {
        return 0, 0, errors.New("Bad result")
    }

    return res.AccountWrapper.StateCurrent.Balance.Value, res.AccountWrapper.StateCurrent.Counter, nil
}

func (c *Collector) globalStatePump() error {
    req := pb.GlobalStateStreamRequest{GlobalStateDataFlags:
        uint32(pb.GlobalStateDataFlag_GLOBAL_STATE_DATA_FLAG_REWARD) |
        uint32(pb.GlobalStateDataFlag_GLOBAL_STATE_DATA_FLAG_TRANSACTION_RECEIPT) |
        uint32(pb.GlobalStateDataFlag_GLOBAL_STATE_DATA_FLAG_ACCOUNT)}

    log.Info("Start global state pump")
    defer func() {
        c.notify <- -streamType_globalState
        log.Info("Stop global state pump")
    }()

    c.notify <- +streamType_globalState

    stream, err := c.globalClient.GlobalStateStream(context.Background(), &req)
    if err != nil {
        log.Error("cannot get global state account stream: %v", err)
        return err
    }

    for {
        response, err := stream.Recv()
        if err == io.EOF {
            return err
        }
        if err != nil {
            log.Error("cannot receive Global state data: %v", err)
            return err
        }
        item := response.GetDatum()
        if account := item.GetAccountWrapper(); account != nil {
            c.listener.OnAccount(account)
        } else if reward := item.GetReward(); reward != nil {
            c.listener.OnReward(reward)
        } else if receipt := item.GetReceipt(); receipt != nil {
            c.listener.OnTransactionReceipt(receipt)
        }
    }

    return nil
}
/*
func (c *Collector) transactionsStatePump() error {
    var req empty.Empty

    log.Info("Start global state transactions state pump")
    defer func() {
        c.notify <- -streamType_global_TransactionState
        log.Info("Stop global state transactions state pump")
    }()

    c.notify <- +streamType_global_TransactionState

    stream, err := c.globalClient.TransactionStateStream(context.Background(), &req)
    if err != nil {
        log.Error("cannot get global state transactions state: %v", err)
        return err
    }

    for {
        txState, err := stream.Recv()
        if err == io.EOF {
            return err
        }
        if err != nil {
            log.Error("cannot receive TransactionState: %v", err)
            return err
        }

        log.Info("TransactionState: %v, %v", txState.GetId(), txState.GetState())
        var id sm.TransactionID
        copy(id[:], txState.GetId().GetId())

        c.history.AddTransactionState(&id, txState.GetState());
    }

    return nil
}
*/
