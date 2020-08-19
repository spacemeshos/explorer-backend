package model

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/explorer-backend/utils"
)

type Account struct {
    Address	string	// account public address
    Balance	uint64	// known account balance
}

type AccountService interface {
    GetAccount(ctx context.Context, query *bson.D) (*Account, error)
    GetAccounts(ctx context.Context, query *bson.D) ([]*Account, error)
    SaveAccount(ctx context.Context, in *Account) error
}

func NewAccount(in *pb.Account) *Account {
    return &Account{
        Address: utils.BytesToAddressString(in.GetAccountId().GetAddress()),
        Balance: in.GetStateCurrent().GetBalance().GetValue(),
    }
}

