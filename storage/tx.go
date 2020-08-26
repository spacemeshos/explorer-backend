package storage

import (
    "context"
    "errors"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/spacemeshos/explorer-backend/model"
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
        return nil, err
    }
    if !cursor.Next(ctx) {
        return nil, errors.New("Empty result")
    }
    doc := cursor.Current
    account := &model.Transaction{
        Id: doc.Lookup("id").String(),
        Layer: uint32(doc.Lookup("layer").Int32()),
        Block: doc.Lookup("block").String(),
        Index: uint32(doc.Lookup("index").Int32()),
        Result: int(doc.Lookup("result").Int32()),
        GasProvided: uint64(doc.Lookup("gasProvided").Int64()),
        GasPrice: uint64(doc.Lookup("gasPrice").Int64()),
        GasUsed: uint64(doc.Lookup("gasUsed").Int64()),
        Fee: uint64(doc.Lookup("fee").Int64()),
        Amount: uint64(doc.Lookup("amount").Int64()),
        Counter: uint64(doc.Lookup("counter").Int64()),
        Type: int(doc.Lookup("type").Int32()),
        Scheme: int(doc.Lookup("scheme").Int32()),
        Signature: doc.Lookup("signature").String(),
        PublicKey: doc.Lookup("pubKey").String(),
        Sender: doc.Lookup("sender").String(),
        Receiver: doc.Lookup("receiver").String(),
        SvmData: doc.Lookup("svmData").String(),
    }
    return account, nil
}

func (s *Storage) GetTransactions(parent context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Transaction, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cursor, err := s.db.Collection("txs").Find(ctx, query, opts...)
    if err != nil {
        return nil, err
    }
    var docs interface{} = []bson.D{}
    err = cursor.All(ctx, &docs)
    if err != nil {
        return nil, err
    }
    if len(docs.([]bson.D)) == 0 {
        return nil, nil
    }
    txs := make([]*model.Transaction, len(docs.([]bson.D)), len(docs.([]bson.D)))
    for i, doc := range docs.([]bson.D) {
        txs[i] = &model.Transaction{
            Id: doc[0].Value.(string),
            Layer: uint32(doc[1].Value.(int32)),
            Block: doc[2].Value.(string),
            Index: uint32(doc[3].Value.(int32)),
            Result: int(doc[4].Value.(int32)),
            GasProvided: uint64(doc[5].Value.(int64)),
            GasPrice: uint64(doc[6].Value.(int64)),
            GasUsed: uint64(doc[7].Value.(int64)),
            Fee: uint64(doc[8].Value.(int64)),
            Amount: uint64(doc[9].Value.(int64)),
            Counter: uint64(doc[10].Value.(int64)),
            Type: int(doc[11].Value.(int32)),
            Scheme: int(doc[12].Value.(int32)),
            Signature: doc[13].Value.(string),
            PublicKey: doc[14].Value.(string),
            Sender: doc[15].Value.(string),
            Receiver: doc[16].Value.(string),
            SvmData: doc[17].Value.(string),
        }
    }
    return txs, nil
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
            return err
        }
    }
    return nil
}

func (s *Storage) UpdateTransaction(parent context.Context, in *model.TransactionReceipt) error {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    _, err := s.db.Collection("txs").UpdateOne(ctx, bson.D{{"id", in.Id}}, bson.D{
        {"index", in.Index},
        {"result", in.Result},
        {"gasUsed", in.GasUsed},
        {"fee", in.Fee},
        {"svmData", in.SvmData},
    })
    return err
}
