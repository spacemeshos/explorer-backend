package model

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/explorer-backend/utils"

    "github.com/spacemeshos/go-spacemesh/crypto/sha3"
)

type Account struct {
    Address	string	// account public address
    Balance	uint64	// known account balance
    Counter	uint64
}

type AccountService interface {
    GetAccount(ctx context.Context, query *bson.D) (*Account, error)
    GetAccounts(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*Account, error)
    SaveAccount(ctx context.Context, in *Account) error
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

