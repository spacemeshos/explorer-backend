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
)

func (s *Storage) GetTransaction(parent context.Context, query *bson.D) (*model.Transaction, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("txs").Find(ctx, query)
    if err != nil {
        return nil, err
    }
    if !cursor.Next(ctx) {
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    account := &model.Transaction{
        Id: doc.Lookup("id").String(),
        Layer: uint64(doc.Lookup("layer").Int64()),
        Block: doc.Lookup("block").String(),
        Index: uint32(doc.Lookup("index").Int32()),
        Result: uint64(doc.Lookup("result").Int()),
        GasProvided: uint64(doc.Lookup("gasProvided").Int64()),
        GasPrice: uint64(doc.Lookup("gasPrice").Int64()),
        GasUsed: uint64(doc.Lookup("gasUsed").Int64()),
        Fee: uint64(doc.Lookup("fee").Int64()),
        Amount: uint64(doc.Lookup("amount").Int64()),
        Counter: uint64(doc.Lookup("counter").Int64()),
        Type: doc.Lookup("type").Int(),
        Scheme: doc.Lookup("scheme").Int(),
        Signature: doc.Lookup("signature").String(),
        PublicKey: doc.Lookup("pubKey").String(),
        Sender: doc.Lookup("sender").String(),
        Receiver: doc.Lookup("receiver").String(),
        SvmData: doc.Lookup("svmData").String(),
    }
    return account, nil
}

func (s *Storage) GetTransactions(parent context.Context, query *bson.D) ([]*model.Transaction, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("txs").Find(ctx, query)
    if err != nil {
        return nil, err
    }
    var docs interface{} = []bson.D{}
    err = cursor.All(ctx, &docs)
    if err != nil {
        return nil, err
    }
    if len(docs) == 0 {
        return nil, nil
    }
    txs := make([]*model.Transaction, len(docs), len(docs))
    for i, doc := range docs {
        txs[i] = &model.Transaction{
            Id: doc.Lookup("id").String(),
            Layer: uint64(doc.Lookup("layer").Int64()),
            Block: doc.Lookup("block").String(),
            Index: uint32(doc.Lookup("index").Int32()),
            Result: uint64(doc.Lookup("result").Int()),
            GasProvided: uint64(doc.Lookup("gasProvided").Int64()),
            GasPrice: uint64(doc.Lookup("gasPrice").Int64()),
            GasUsed: uint64(doc.Lookup("gasUsed").Int64()),
            Fee: uint64(doc.Lookup("fee").Int64()),
            Amount: uint64(doc.Lookup("amount").Int64()),
            Counter: uint64(doc.Lookup("counter").Int64()),
            Type: doc.Lookup("type").Int(),
            Scheme: doc.Lookup("scheme").Int(),
            Signature: doc.Lookup("signature").String(),
            PublicKey: doc.Lookup("pubKey").String(),
            Sender: doc.Lookup("sender").String(),
            Receiver: doc.Lookup("receiver").String(),
            SvmData: doc.Lookup("svmData").String(),
        }
    }
    return txs, nil
}

func (s *Storage) SaveTransaction(parent context.Context, in *model.Transaction) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    res, err := s.db.Collection("txs").InsertOne(ctx, bson.D{
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
    return err
}

func (s *Storage) SaveTransactions(parent context.Context, in map[string]*model.Transaction) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    for _, tx := range in {
        res, err := s.db.Collection("txs").InsertOne(ctx, bson.D{
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
            return err
        }
    }
    return nil
}

func (s *Storage) UpdateTransaction(parent context.Context, in *model.TransactionReceipt) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    res, err := s.db.Collection("txs").UpdateOne(ctx, bson.D{{{"id", in.Id}}}, bson.D{
        {"index", in.Index},
        {"result", in.Result},
        {"gasUsed", in.GasUsed},
        {"fee", in.Fee},
        {"svmData", in.SvmData},
    })
    return err
}
