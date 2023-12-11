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

func (s *Storage) InitTransactionsStorage(ctx context.Context) error {
	models := []mongo.IndexModel{
		{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetName("idIndex").SetUnique(true)},
		{Keys: bson.D{{Key: "layer", Value: 1}}, Options: options.Index().SetName("layerIndex").SetUnique(false)},
		{Keys: bson.D{{Key: "block", Value: 1}}, Options: options.Index().SetName("blockIndex").SetUnique(false)},
		{Keys: bson.D{{Key: "sender", Value: 1}}, Options: options.Index().SetName("senderIndex").SetUnique(false)},
		{Keys: bson.D{{Key: "receiver", Value: 1}}, Options: options.Index().SetName("receiverIndex").SetUnique(false)},
		{Keys: bson.D{{Key: "timestamp", Value: -1}}, Options: options.Index().SetName("timestampIndex").SetUnique(false)},
		{Keys: bson.D{{Key: "counter", Value: -1}}, Options: options.Index().SetName("counterIndex").SetUnique(false)},
	}
	_, err := s.db.Collection("txs").Indexes().CreateMany(ctx, models, options.CreateIndexes().SetMaxTime(20*time.Second))
	return err
}

func (s *Storage) GetTransaction(parent context.Context, query *bson.D) (*model.Transaction, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cursor, err := s.db.Collection("txs").Find(ctx, query)
	if err != nil {
		log.Info("GetTransaction: %v", err)
		return nil, err
	}
	if !cursor.Next(ctx) {
		log.Info("GetTransaction: Empty result")
		return nil, errors.New("empty result")
	}
	doc := cursor.Current
	tx := &model.Transaction{
		Id:         utils.GetAsString(doc.Lookup("id")),
		Layer:      utils.GetAsUInt32(doc.Lookup("layer")),
		Block:      utils.GetAsString(doc.Lookup("block")),
		BlockIndex: utils.GetAsUInt32(doc.Lookup("blockIndex")),
		Index:      utils.GetAsUInt32(doc.Lookup("index")),
		State:      utils.GetAsInt(doc.Lookup("state")),
		Timestamp:  utils.GetAsUInt32(doc.Lookup("timestamp")),
		MaxGas:     utils.GetAsUInt64(doc.Lookup("maxGas")),
		GasPrice:   utils.GetAsUInt64(doc.Lookup("gasPrice")),
		GasUsed:    utils.GetAsUInt64(doc.Lookup("gasUsed")),
		Fee:        utils.GetAsUInt64(doc.Lookup("fee")),
		Amount:     utils.GetAsUInt64(doc.Lookup("amount")),
		Counter:    utils.GetAsUInt64(doc.Lookup("counter")),
		Type:       utils.GetAsInt(doc.Lookup("type")),
		Signature:  utils.GetAsString(doc.Lookup("signature")),
		PublicKey:  utils.GetAsString(doc.Lookup("pubKey")),
		Sender:     utils.GetAsString(doc.Lookup("sender")),
		Receiver:   utils.GetAsString(doc.Lookup("receiver")),
		SvmData:    utils.GetAsString(doc.Lookup("svmData")),
	}
	return tx, nil
}

func (s *Storage) GetTransactionsCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) int64 {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	count, err := s.db.Collection("txs").CountDocuments(ctx, query, opts...)
	if err != nil {
		log.Info("GetTransactionsCount: %v", err)
		return 0
	}
	return count
}

func (s *Storage) GetTransactionsAmount(parent context.Context, query *bson.D) int64 {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	matchStage := bson.D{
		{Key: "$match", Value: query},
	}
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: ""},
			{Key: "amount", Value: bson.D{
				{Key: "$sum", Value: "$amount"},
			}},
		}},
	}
	cursor, err := s.db.Collection("txs").Aggregate(ctx, mongo.Pipeline{
		matchStage,
		groupStage,
	})
	if err != nil {
		log.Info("GetTransactionsAmount: %v", err)
		return 0
	}
	if !cursor.Next(ctx) {
		log.Info("GetTransactionsAmount: Empty result")
		return 0
	}
	doc := cursor.Current
	return utils.GetAsInt64(doc.Lookup("amount"))
}

func (s *Storage) IsTransactionExists(parent context.Context, txId string) bool {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	count, err := s.db.Collection("txs").CountDocuments(ctx, bson.D{{Key: "id", Value: txId}})
	if err != nil {
		log.Info("IsTransactionExists: %v", err)
		return false
	}
	return count > 0
}

func (s *Storage) GetTransactions(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]bson.D, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cursor, err := s.db.Collection("txs").Find(ctx, query, opts...)
	if err != nil {
		log.Info("GetTransactions: %v", err)
		return nil, err
	}
	var docs interface{} = []bson.D{}
	err = cursor.All(ctx, &docs)
	if err != nil {
		log.Info("GetTransactions: %v", err)
		return nil, err
	}
	if len(docs.([]bson.D)) == 0 {
		return nil, nil
	}
	return docs.([]bson.D), nil
}

func (s *Storage) SaveTransaction(parent context.Context, in *model.Transaction) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	transaction, err := s.GetTransaction(ctx, &bson.D{{Key: "id", Value: in.Id}})
	if err != nil && err.Error() != "empty result" {
		return err
	}

	tx := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{Key: "id", Value: in.Id},
				{Key: "layer", Value: in.Layer},
				{Key: "block", Value: in.Block},
				{Key: "blockIndex", Value: in.BlockIndex},
				{Key: "index", Value: in.Index},
				{Key: "state", Value: in.State},
				{Key: "timestamp", Value: in.Timestamp},
				{Key: "maxGas", Value: in.MaxGas},
				{Key: "gasPrice", Value: in.GasPrice},
				{Key: "gasUsed", Value: in.GasUsed},
				{Key: "fee", Value: in.Fee},
				{Key: "amount", Value: in.Amount},
				{Key: "counter", Value: in.Counter},
				{Key: "type", Value: in.Type},
				{Key: "signature", Value: in.Signature},
				{Key: "pubKey", Value: in.PublicKey},
				{Key: "sender", Value: in.Sender},
				{Key: "receiver", Value: in.Receiver},
				{Key: "svmData", Value: in.SvmData},
				{Key: "message", Value: in.Message},
				{Key: "touchedAddresses", Value: in.TouchedAddresses},
			},
		},
	}

	if transaction != nil {
		tx = bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{Key: "id", Value: in.Id},
					{Key: "layer", Value: in.Layer},
					{Key: "block", Value: in.Block},
					{Key: "blockIndex", Value: in.BlockIndex},
					{Key: "index", Value: in.Index},
					{Key: "timestamp", Value: in.Timestamp},
					{Key: "maxGas", Value: in.MaxGas},
					{Key: "gasPrice", Value: in.GasPrice},
					{Key: "fee", Value: in.Fee},
					{Key: "amount", Value: in.Amount},
					{Key: "counter", Value: in.Counter},
					{Key: "type", Value: in.Type},
					{Key: "signature", Value: in.Signature},
					{Key: "pubKey", Value: in.PublicKey},
					{Key: "sender", Value: in.Sender},
					{Key: "receiver", Value: in.Receiver},
					{Key: "svmData", Value: in.SvmData},
				},
			},
		}
	}

	_, err = s.db.Collection("txs").UpdateOne(ctx,
		bson.D{{Key: "id", Value: in.Id}}, tx, options.Update().SetUpsert(true))
	if err != nil {
		log.Info("SaveTransaction: %v obj: %+v", err, tx)
	}
	return err
}

func (s *Storage) SaveTransactions(parent context.Context, in map[string]*model.Transaction) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	for _, tx := range in {
		_, err := s.db.Collection("txs").InsertOne(ctx, bson.D{
			{Key: "id", Value: tx.Id},
			{Key: "layer", Value: tx.Layer},
			{Key: "block", Value: tx.Block},
			{Key: "blockIndex", Value: tx.BlockIndex},
			{Key: "index", Value: tx.Index},
			{Key: "state", Value: tx.State},
			{Key: "timestamp", Value: tx.Timestamp},
			{Key: "maxGas", Value: tx.MaxGas},
			{Key: "gasPrice", Value: tx.GasPrice},
			{Key: "gasUsed", Value: tx.GasUsed},
			{Key: "fee", Value: tx.Fee},
			{Key: "amount", Value: tx.Amount},
			{Key: "counter", Value: tx.Counter},
			{Key: "type", Value: tx.Type},
			{Key: "signature", Value: tx.Signature},
			{Key: "pubKey", Value: tx.PublicKey},
			{Key: "sender", Value: tx.Sender},
			{Key: "receiver", Value: tx.Receiver},
			{Key: "svmData", Value: tx.SvmData},
		})
		if err != nil {
			log.Info("SaveTransactions: %v", err)
			return err
		}
	}
	return nil
}

func (s *Storage) SaveTransactionResult(parent context.Context, in *model.Transaction) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	transaction, err := s.GetTransaction(ctx, &bson.D{{Key: "id", Value: in.Id}})
	if err != nil && err.Error() != "empty result" {
		return err
	}

	tx := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{Key: "id", Value: in.Id},
				{Key: "layer", Value: in.Layer},
				{Key: "block", Value: in.Block},
				{Key: "blockIndex", Value: in.BlockIndex},
				{Key: "index", Value: in.Index},
				{Key: "state", Value: in.State},
				{Key: "timestamp", Value: in.Timestamp},
				{Key: "maxGas", Value: in.MaxGas},
				{Key: "gasPrice", Value: in.GasPrice},
				{Key: "gasUsed", Value: in.GasUsed},
				{Key: "fee", Value: in.Fee},
				{Key: "amount", Value: in.Amount},
				{Key: "counter", Value: in.Counter},
				{Key: "type", Value: in.Type},
				{Key: "signature", Value: in.Signature},
				{Key: "pubKey", Value: in.PublicKey},
				{Key: "sender", Value: in.Sender},
				{Key: "receiver", Value: in.Receiver},
				{Key: "svmData", Value: in.SvmData},
				{Key: "message", Value: in.Message},
				{Key: "touchedAddresses", Value: in.TouchedAddresses},
			},
		},
	}

	if transaction != nil {
		tx = bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{Key: "id", Value: in.Id},
					{Key: "state", Value: in.State},
					{Key: "gasUsed", Value: in.GasUsed},
					{Key: "fee", Value: in.Fee},
					{Key: "message", Value: in.Message},
					{Key: "touchedAddresses", Value: in.TouchedAddresses},
				},
			},
		}
	}

	_, err = s.db.Collection("txs").UpdateOne(ctx,
		bson.D{{Key: "id", Value: in.Id}}, tx, options.Update().SetUpsert(true))
	if err != nil {
		log.Info("SaveTransactionResult: %v obj: %+v", err, tx)
	}
	return err
}
