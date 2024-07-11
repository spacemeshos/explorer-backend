package storage

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/go-spacemesh/log"

	"github.com/spacemeshos/explorer-backend/model"
	"github.com/spacemeshos/explorer-backend/utils"
)

func (s *Storage) InitAccountsStorage(ctx context.Context) error {
	if _, err := s.db.Collection("accounts").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "address", Value: 1}},
		Options: options.Index().SetName("addressIndex").SetUnique(true)}); err != nil {
		return err
	}

	if _, err := s.db.Collection("accounts").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "created", Value: 1}},
		Options: options.Index().SetName("createIndex").SetUnique(false)}); err != nil {
		return err
	}

	if _, err := s.db.Collection("accounts").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "layer", Value: -1}},
		Options: options.Index().SetName("modifiedIndex").SetUnique(false)}); err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetAccount(parent context.Context, query *bson.D) (*model.Account, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cursor, err := s.db.Collection("accounts").Find(ctx, query)
	if err != nil {
		log.Info("GetAccount: %v", err)
		return nil, err
	}
	if !cursor.Next(ctx) {
		log.Info("GetAccount: Empty result", err)
		return nil, errors.New("Empty result")
	}
	doc := cursor.Current
	account := &model.Account{
		Address: utils.GetAsString(doc.Lookup("address")),
		Balance: utils.GetAsUInt64(doc.Lookup("balance")),
		Counter: utils.GetAsUInt64(doc.Lookup("counter")),
	}
	return account, nil
}

func (s *Storage) GetAccountsCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) int64 {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	count, err := s.db.Collection("accounts").CountDocuments(ctx, query, opts...)
	if err != nil {
		log.Info("GetAccountsCount: %v", err)
		return 0
	}
	return count
}

func (s *Storage) GetAccounts(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cursor, err := s.db.Collection("accounts").Find(ctx, query, opts...)
	if err != nil {
		log.Info("GetAccounts: %v", err)
		return nil, err
	}
	var docs interface{} = []bson.D{}
	err = cursor.All(ctx, &docs)
	if err != nil {
		log.Info("GetAccounts: %v", err)
		return nil, err
	}
	if len(docs.([]bson.D)) == 0 {
		log.Info("GetAccounts: Empty result")
		return nil, nil
	}
	return docs.([]bson.D), nil
}

func (s *Storage) AddAccount(parent context.Context, layer uint32, address string, balance uint64) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	acc := bson.D{
		{Key: "$set",
			Value: bson.D{
				{Key: "address", Value: address},
				{Key: "layer", Value: layer},
				{Key: "balance", Value: balance},
				{Key: "counter", Value: uint64(0)},
			},
		},
		{Key: "$setOnInsert",
			Value: bson.D{
				{Key: "created", Value: layer},
			},
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := s.db.Collection("accounts").UpdateOne(ctx, bson.D{{Key: "address", Value: address}}, acc, opts)
	if err != nil {
		log.Info("AddAccount: %v", err)
	}
	return nil
}

func (s *Storage) AddAccountQuery(layer uint32, address string, balance uint64) *mongo.UpdateOneModel {
	filter := bson.D{{Key: "address", Value: address}}
	acc := bson.D{
		{Key: "$set",
			Value: bson.D{
				{Key: "address", Value: address},
				{Key: "layer", Value: layer},
				{Key: "balance", Value: balance},
				{Key: "counter", Value: uint64(0)},
			},
		},
		{Key: "$setOnInsert",
			Value: bson.D{
				{Key: "created", Value: layer},
			},
		},
	}

	accountModel := mongo.NewUpdateOneModel()
	accountModel.SetFilter(filter)
	accountModel.SetUpdate(acc)
	accountModel.SetUpsert(true)

	return accountModel
}

func (s *Storage) SaveAccount(parent context.Context, layer uint32, in *model.Account) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	_, err := s.db.Collection("accounts").UpdateOne(ctx, bson.D{{Key: "address", Value: in.Address}}, bson.D{
		{Key: "$set",
			Value: bson.D{
				{Key: "address", Value: in.Address},
				{Key: "layer", Value: layer},
				{Key: "balance", Value: in.Balance},
				{Key: "counter", Value: in.Counter},
			}},
		{Key: "$setOnInsert",
			Value: bson.D{
				{Key: "created", Value: layer},
			},
		},
	}, options.Update().SetUpsert(true))
	if err != nil {
		log.Info("SaveAccount: %v", err)
	}
	return nil
}

func (s *Storage) UpdateAccount(parent context.Context, address string, balance uint64, counter uint64) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	_, err := s.db.Collection("accounts").UpdateOne(ctx, bson.D{{Key: "address", Value: address}}, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "balance", Value: balance},
			{Key: "counter", Value: counter},
		}},
	}, options.Update().SetUpsert(true))
	if err != nil {
		log.Info("UpdateAccount: %v", err)
	}
	return nil
}

func (s *Storage) AddAccountSent(parent context.Context, layer uint32, address string, amount uint64, fee uint64) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	_, err := s.db.Collection("accounts").UpdateOne(ctx, bson.D{{Key: "address", Value: address}}, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "layer", Value: layer},
		}},
	})
	if err != nil {
		log.Info("AddAccountSent: update account touch error %v", err)
	}
	return nil
}

func (s *Storage) AddAccountReceived(parent context.Context, layer uint32, address string, amount uint64) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	_, err := s.db.Collection("accounts").UpdateOne(ctx, bson.D{{Key: "address", Value: address}}, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "layer", Value: layer},
		}},
	})
	if err != nil {
		log.Info("AddAccountReceived: update account touch error %v", err)
	}
	return nil
}

func (s *Storage) AddAccountReward(parent context.Context, layer uint32, address string, reward uint64, fee uint64) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	_, err := s.db.Collection("accounts").UpdateOne(ctx, bson.D{{Key: "address", Value: address}}, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "layer", Value: layer},
		}},
	})
	if err != nil {
		log.Info("AddAccountReward: update account touch error %v", err)
	}
	return nil
}
