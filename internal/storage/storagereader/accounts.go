package storagereader

import (
	"context"
	"fmt"
	bson "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// CountAccounts returns the number of accounts matching the query.
func (s *Reader) CountAccounts(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	return s.db.Collection("accounts").CountDocuments(ctx, query, opts...)
}

// GetAccounts returns the accounts matching the query.
func (s *Reader) GetAccounts(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Account, error) {
	cursor, err := s.db.Collection("accounts").Find(ctx, query, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	var docs []*model.Account
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("error decode accounts: %w", err)
	}

	for _, doc := range docs {
		summary, err := s.GetAccountSummary(ctx, doc.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to get account summary: %w", err)
		}
		if summary == nil {
			continue
		}

		doc.Balance = summary.Balance
		doc.Sent = summary.Sent
		doc.Received = summary.Received
		doc.Awards = summary.Awards
		doc.Fees = summary.Fees
		doc.LayerTms = int32(s.GetLayerTimestamp(uint32(doc.Created)))
	}
	return docs, nil
}

// GetAccountSummary returns the summary of the accounts matching the query. Not all accounts from api have filled this data.
func (s *Reader) GetAccountSummary(ctx context.Context, address string) (*model.AccountSummary, error) {
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "address", Value: address}}}},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "txs"},
					{Key: "localField", Value: "address"},
					{Key: "foreignField", Value: "receiver"},
					{Key: "as", Value: "txs_received"},
				},
			},
		},
		bson.D{
			{Key: "$unwind",
				Value: bson.D{
					{Key: "path", Value: "$txs_received"},
					{Key: "preserveNullAndEmptyArrays", Value: true},
				},
			},
		},
		bson.D{
			{Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: "$_id"},
					{Key: "received", Value: bson.D{{Key: "$sum", Value: "$txs_received.amount"}}},
					{Key: "address", Value: bson.D{{Key: "$first", Value: "$$ROOT.address"}}},
				},
			},
		},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "txs"},
					{Key: "localField", Value: "address"},
					{Key: "foreignField", Value: "sender"},
					{Key: "as", Value: "txs_sent"},
				},
			},
		},
		bson.D{
			{Key: "$unwind",
				Value: bson.D{
					{Key: "path", Value: "$txs_sent"},
					{Key: "preserveNullAndEmptyArrays", Value: true},
				},
			},
		},
		bson.D{
			{Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: "$_id"},
					{Key: "address", Value: bson.D{{Key: "$first", Value: "$$ROOT.address"}}},
					{Key: "received", Value: bson.D{{Key: "$first", Value: "$$ROOT.received"}}},
					{Key: "sent", Value: bson.D{{Key: "$sum", Value: "$txs_sent.amount"}}},
					{Key: "fees", Value: bson.D{{Key: "$sum", Value: "$txs_sent.fee"}}},
				},
			},
		},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "rewards"},
					{Key: "localField", Value: "address"},
					{Key: "foreignField", Value: "coinbase"},
					{Key: "as", Value: "rewards"},
				},
			},
		},
		bson.D{
			{Key: "$unwind",
				Value: bson.D{
					{Key: "path", Value: "$rewards"},
					{Key: "preserveNullAndEmptyArrays", Value: true},
				},
			},
		},
		bson.D{
			{Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: ""},
					{Key: "sent", Value: bson.D{{Key: "$first", Value: "$$ROOT.sent"}}},
					{Key: "received", Value: bson.D{{Key: "$first", Value: "$$ROOT.received"}}},
					{Key: "fees", Value: bson.D{{Key: "$first", Value: "$$ROOT.fees"}}},
					{Key: "awards", Value: bson.D{{Key: "$sum", Value: "$rewards.total"}}},
					{Key: "layer", Value: bson.D{{Key: "$max", Value: "$rewards.layer"}}},
				},
			},
		},
		bson.D{
			{Key: "$addFields",
				Value: bson.D{
					{Key: "balance",
						Value: bson.D{
							{Key: "$add",
								Value: bson.A{
									"$awards",
									"$received",
								},
							},
						},
					},
				},
			},
		},
		bson.D{
			{Key: "$addFields",
				Value: bson.D{
					{Key: "balance",
						Value: bson.D{
							{Key: "$subtract",
								Value: bson.A{
									"$balance",
									bson.D{
										{Key: "$add",
											Value: bson.A{
												"$fees",
												"$sent",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		bson.D{
			{Key: "$addFields",
				Value: bson.D{
					{Key: "balance",
						Value: bson.D{
							{Key: "$max",
								Value: bson.A{
									0,
									"$balance",
								},
							},
						},
					},
				},
			},
		},
	}

	cursor, err := s.db.Collection("accounts").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error get account summary: %w", err)
	}
	if !cursor.Next(ctx) {
		return nil, nil
	}
	var accSummary model.AccountSummary
	if err = cursor.Decode(&accSummary); err != nil {
		return nil, fmt.Errorf("error decode account summary: %w", err)
	}
	return &accSummary, nil
}
