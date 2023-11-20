package storagereader

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spacemeshos/explorer-backend/model"
)

// CountAccounts returns the number of accounts matching the query.
func (s *Reader) CountAccounts(ctx context.Context, query *bson.D, opts ...*options.CountOptions) (int64, error) {
	return s.db.Collection("accounts").CountDocuments(ctx, query, opts...)
}

// GetAccounts returns the accounts matching the query.
func (s *Reader) GetAccounts(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*model.Account, error) {
	skip := int64(0)
	if opts[0].Skip != nil {
		skip = *opts[0].Skip
	}

	pipeline := bson.A{
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "txs"},
					{"let", bson.D{{"addr", "$address"}}},
					{"pipeline",
						bson.A{
							bson.D{
								{"$match",
									bson.D{
										{"$or",
											bson.A{
												bson.D{
													{"$expr",
														bson.D{
															{"$eq",
																bson.A{
																	"$sender",
																	"$$addr",
																},
															},
														},
													},
												},
												bson.D{
													{"$expr",
														bson.D{
															{"$eq",
																bson.A{
																	"$receiver",
																	"$$addr",
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
							bson.D{{"$sort", bson.D{{"layer", 1}}}},
							bson.D{{"$limit", 1}},
							bson.D{
								{"$project",
									bson.D{
										{"_id", 0},
										{"layer", 1},
									},
								},
							},
						},
					},
					{"as", "createdLayerRst"},
				},
			},
		},
		bson.D{
			{"$addFields",
				bson.D{
					{"createdLayer",
						bson.D{
							{"$arrayElemAt",
								bson.A{
									"$createdLayerRst.layer",
									0,
								},
							},
						},
					},
				},
			},
		},
		bson.D{{"$project", bson.D{{"createdLayerRst", 0}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "createdLayer", Value: -1}}}},
		bson.D{{Key: "$skip", Value: skip}},
		bson.D{{Key: "$limit", Value: *opts[0].Limit}},
	}

	if query != nil {
		pipeline = append(bson.A{
			bson.D{{Key: "$match", Value: *query}},
		}, pipeline...)
	}

	cursor, err := s.db.Collection("accounts").Aggregate(ctx, pipeline)
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

		doc.Sent = summary.Sent
		doc.Received = summary.Received
		doc.Awards = summary.Awards
		doc.Fees = summary.Fees
		doc.LastActivity = summary.LastActivity
	}
	return docs, nil
}

// GetAccountSummary returns the summary of the accounts matching the query. Not all accounts from api have filled this data.
func (s *Reader) GetAccountSummary(ctx context.Context, address string) (*model.AccountSummary, error) {
	var accSummary model.AccountSummary

	totalRewards, _, err := s.CountCoinbaseRewards(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("error occured while getting sum of rewards: %w", err)
	}
	accSummary.Awards = uint64(totalRewards)

	received, _, err := s.CountReceivedTransactions(ctx, address)
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("error occured while getting sum of received txs: %w", err)
		}
	}
	accSummary.Received = uint64(received)

	sent, fees, _, err := s.CountSentTransactions(ctx, address)
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("error occured while getting sum of sent txs: %w", err)
		}
	}
	accSummary.Sent = uint64(sent)
	accSummary.Fees = uint64(fees)

	latestTx, err := s.GetLatestTransaction(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("error occured while getting latest sent txs: %w", err)
	}

	latestReward, err := s.GetLatestReward(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("error occured while getting latest reawrd: %w", err)
	}

	if latestTx != nil {
		if latestReward != nil && latestReward.Layer > latestTx.Layer {
			accSummary.LastActivity = int32(s.GetLayerTimestamp(latestReward.Layer))
		} else {
			accSummary.LastActivity = int32(s.GetLayerTimestamp(latestTx.Layer))
		}
	} else {
		if latestReward != nil {
			accSummary.LastActivity = int32(s.GetLayerTimestamp(latestReward.Layer))
		}
	}

	return &accSummary, nil
}
