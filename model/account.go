package model

import (
	"context"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/crypto/sha3"

	"github.com/spacemeshos/explorer-backend/utils"
)

type Account struct {
	Address string `json:"address" bson:"address"` // account public address
	Balance uint64 `json:"balance" bson:"balance"` // known account balance
	Counter uint64 `json:"counter" bson:"counter"`
	Created uint64 `json:"created" bson:"created"`
	// get from ledger collection
	Sent     uint64 `json:"sent" bson:"-"`
	Received uint64 `json:"received" bson:"-"`
	Awards   uint64 `json:"awards" bson:"-"`
	Fees     uint64 `json:"fees" bson:"-"`
	LayerTms int32  `json:"timestamp" bson:"-"`
	Txs      int64  `json:"txs" bson:"-"`
}

// AccountSummary data taken from `ledger` collection. Not all accounts from api have filled this data.
type AccountSummary struct {
	Sent     uint64 `json:"sent" bson:"sent"`
	Received uint64 `json:"received" bson:"received"`
	Awards   uint64 `json:"awards" bson:"awards"`
	Fees     uint64 `json:"fees" bson:"fees"`
	LayerTms int32  `json:"timestamp" bson:"timestamp"`
}

type AccountService interface {
	GetAccount(ctx context.Context, accountID string) (*Account, error)
	GetAccounts(ctx context.Context, page, perPage int64) ([]*Account, int64, error)
	GetAccountTransactions(ctx context.Context, accountID string, page, perPage int64) ([]*Transaction, int64, error)
	GetAccountRewards(ctx context.Context, accountID string, page, perPage int64) ([]*Reward, int64, error)
}

func NewAccount(in *pb.Account) *Account {
	return &Account{
		Address: utils.BytesToAddressString(in.GetAccountId().GetAddress()),
		Balance: in.GetStateCurrent().GetBalance().GetValue(),
		Counter: in.GetStateCurrent().GetCounter(),
	}
}

// Hex returns an EIP55-compliant hex string representation of the address.
func ToCheckedAddress(a string) string {
	if len(a) != 42 || a[0] != '0' || a[1] != 'x' {
		return ""
	}
	unchecksummed := make([]byte, 40)
	copy(unchecksummed, a[2:])
	sha := sha3.NewKeccak256()
	sha.Write([]byte(unchecksummed))
	hash := sha.Sum(nil)

	result := []byte(unchecksummed)
	for i := 0; i < len(result); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}
	return "0x" + string(result)
}
