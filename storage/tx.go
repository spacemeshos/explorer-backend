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
        {Keys: bson.D{{"id", 1}}, Options: options.Index().SetName("idIndex").SetUnique(true)},
        {Keys: bson.D{{"layer", 1}}, Options: options.Index().SetName("layerIndex").SetUnique(false)},
        {Keys: bson.D{{"block", 1}}, Options: options.Index().SetName("blockIndex").SetUnique(false)},
        {Keys: bson.D{{"sender", 1}}, Options: options.Index().SetName("senderIndex").SetUnique(false)},
        {Keys: bson.D{{"receiver", 1}}, Options: options.Index().SetName("receiverIndex").SetUnique(false)},
    }
    _, err := s.db.Collection("txs").Indexes().CreateMany(ctx, models, options.CreateIndexes().SetMaxTime(2 * time.Second));
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
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    account := &model.Transaction{
        Id: doc.Lookup("id").String(),
        Layer: utils.GetAsUInt32(doc.Lookup("layer")),
        Block: doc.Lookup("block").String(),
        Index: utils.GetAsUInt32(doc.Lookup("index")),
        Result: utils.GetAsInt(doc.Lookup("result")),
        GasProvided: utils.GetAsUInt64(doc.Lookup("gasProvided")),
        GasPrice: utils.GetAsUInt64(doc.Lookup("gasPrice")),
        GasUsed: utils.GetAsUInt64(doc.Lookup("gasUsed")),
        Fee: utils.GetAsUInt64(doc.Lookup("fee")),
        Amount: utils.GetAsUInt64(doc.Lookup("amount")),
        Counter: utils.GetAsUInt64(doc.Lookup("counter")),
        Type: utils.GetAsInt(doc.Lookup("type")),
        Scheme: utils.GetAsInt(doc.Lookup("scheme")),
        Signature: doc.Lookup("signature").String(),
        PublicKey: doc.Lookup("pubKey").String(),
        Sender: doc.Lookup("sender").String(),
        Receiver: doc.Lookup("receiver").String(),
        SvmData: doc.Lookup("svmData").String(),
    }
    return account, nil
}

func (s *Storage) GetTransactionsCount(parent context.Context, query *bson.D, opts ...*options.CountOptions) int64 {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    count, err := s.db.Collection("transactions").CountDocuments(ctx, query, opts...)
    if err != nil {
        log.Info("GetTransactionsCount: %v", err)
        return 0
    }
    return count
}

func (s *Storage) IsTransactionExists(parent context.Context, txId string) bool {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    count, err := s.db.Collection("transactions").CountDocuments(ctx, bson.D{{"id", txId}})
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
    _, err := s.db.Collection("txs").InsertOne(ctx, bson.D{
        {"id", in.Id},
        {"layer", in.Layer},
        {"block", in.Block},
        {"index", in.Index},
        {"result", in.Result},
        {"gasProvided", in.GasProvided},
        {"gasPrice", in.GasPrice},
        {"gasUsed", in.GasUsed},
        {"fee", in.Fee},
        {"amount", in.Amount},
        {"counter", in.Counter},
        {"type", in.Type},
        {"scheme", in.Scheme},
        {"signature", in.Signature},
        {"pubKey", in.PublicKey},
        {"sender", in.Sender},
        {"receiver", in.Receiver},
        {"svmData", in.SvmData},
    })
    if err != nil {
        log.Info("SaveTransaction: %v", err)
    }
    return err
}

func (s *Storage) SaveTransactions(parent context.Context, in map[string]*model.Transaction) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    for _, tx := range in {
        _, err := s.db.Collection("txs").InsertOne(ctx, bson.D{
            {"id", tx.Id},
            {"layer", tx.Layer},
            {"block", tx.Block},
            {"index", tx.Index},
            {"result", tx.Result},
            {"gasProvided", tx.GasProvided},
            {"gasPrice", tx.GasPrice},
            {"gasUsed", tx.GasUsed},
            {"fee", tx.Fee},
            {"amount", tx.Amount},
            {"counter", tx.Counter},
            {"type", tx.Type},
            {"scheme", tx.Scheme},
            {"signature", tx.Signature},
            {"pubKey", tx.PublicKey},
            {"sender", tx.Sender},
            {"receiver", tx.Receiver},
            {"svmData", tx.SvmData},
        })
        if err != nil {
            log.Info("SaveTransactions: %v", err)
            return err
        }
    }
    return nil
}

func (s *Storage) UpdateTransaction(parent context.Context, in *model.TransactionReceipt) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("txs").UpdateOne(ctx, bson.D{{"id", in.Id}}, bson.D{
        {"$set", bson.D{
            {"index", in.Index},
            {"result", in.Result},
            {"gasUsed", in.GasUsed},
            {"fee", in.Fee},
            {"svmData", in.SvmData},
        }},
    })
    if err != nil {
        log.Info("UpdateTransaction: %v", err)
    }
    return err
}
