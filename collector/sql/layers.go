package sql

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/sql"
	"github.com/spacemeshos/go-spacemesh/sql/atxs"
	"github.com/spacemeshos/go-spacemesh/sql/ballots"
	"github.com/spacemeshos/go-spacemesh/sql/blocks"
	"github.com/spacemeshos/go-spacemesh/sql/layers"
	"github.com/spacemeshos/go-spacemesh/sql/transactions"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *Client) GetLayer(db *sql.Database, lid types.LayerID, numLayers uint32) (*pb.Layer, error) {
	var bs []*pb.Block
	var activations []types.ATXID

	blts, err := ballots.Layer(db, lid)
	if err != nil {
		return nil, err
	}

	blks, err := blocks.Layer(db, lid)
	if err != nil {
		return nil, err
	}

	layer := types.NewExistingLayer(lid, blts, blks)

	for _, b := range layer.Blocks() {
		mtxs, missing := getMeshTransactions(db, b.TxIDs)
		if len(missing) != 0 {
			return nil, status.Errorf(codes.Internal, "error retrieving tx data")
		}

		pbTxs := make([]*pb.Transaction, 0, len(mtxs))
		for _, t := range mtxs {
			pbTxs = append(pbTxs, castTransaction(&t.Transaction))
		}

		bs = append(bs, &pb.Block{
			Id:           types.Hash20(b.ID()).Bytes(),
			Transactions: pbTxs,
		})
	}

	epoch := lid.Uint32() / numLayers
	if lid.Uint32()%numLayers == 0 {
		atxsId, err := atxs.GetIDsByEpoch(context.Background(), db, types.EpochID(epoch-1))
		if err != nil {
			return nil, err
		}

		activations = append(activations, atxsId...)
	}

	// Extract ATX data from block data
	var pbActivations []*pb.Activation

	// Add unique ATXIDs
	atxids, matxs := GetATXs(db, activations)
	if len(matxs) != 0 {
		return nil, status.Errorf(codes.Internal, "error retrieving activations data")
	}
	for _, atx := range atxids {
		pbActivations = append(pbActivations, convertActivation(atx))
	}

	stateRoot, err := layers.GetStateHash(db, layer.Index())
	if err != nil {
		// This is expected. We can only retrieve state root for a layer that was applied to state,
		// which only happens after it's approved/confirmed.
		log.Debug("no state root for layer", err)
	}

	hash, err := layers.GetAggregatedHash(db, lid)
	if err != nil {
		// This is expected. We can only retrieve state root for a layer that was applied to state,
		// which only happens after it's approved/confirmed.
		log.Debug("no mesh hash at layer", err)
	}
	return &pb.Layer{
		Number:        &pb.LayerNumber{Number: layer.Index().Uint32()},
		Status:        pb.Layer_LAYER_STATUS_CONFIRMED,
		Blocks:        bs,
		Activations:   pbActivations,
		Hash:          hash.Bytes(),
		RootStateHash: stateRoot.Bytes(),
	}, nil
}

func getMeshTransactions(db *sql.Database, ids []types.TransactionID) ([]*types.MeshTransaction, map[types.TransactionID]struct{}) {
	if ids == nil {
		return []*types.MeshTransaction{}, map[types.TransactionID]struct{}{}
	}
	missing := make(map[types.TransactionID]struct{})
	mtxs := make([]*types.MeshTransaction, 0, len(ids))
	for _, tid := range ids {
		var (
			mtx *types.MeshTransaction
			err error
		)
		if mtx, err = transactions.Get(db, tid); err != nil {
			missing[tid] = struct{}{}
		} else {
			mtxs = append(mtxs, mtx)
		}
	}
	return mtxs, missing
}

func GetATXs(db *sql.Database, atxIds []types.ATXID) (map[types.ATXID]*types.VerifiedActivationTx, []types.ATXID) {
	var mIds []types.ATXID
	a := make(map[types.ATXID]*types.VerifiedActivationTx, len(atxIds))
	for _, id := range atxIds {
		t, err := getFullAtx(db, id)
		if err != nil {
			mIds = append(mIds, id)
		} else {
			a[t.ID()] = t
		}
	}
	return a, mIds
}

func getFullAtx(db *sql.Database, id types.ATXID) (*types.VerifiedActivationTx, error) {
	if id == types.EmptyATXID {
		return nil, errors.New("trying to fetch empty atx id")
	}

	atx, err := atxs.Get(db, id)
	if err != nil {
		return nil, fmt.Errorf("get ATXs from DB: %w", err)
	}

	return atx, nil
}

func castTransaction(t *types.Transaction) *pb.Transaction {
	tx := &pb.Transaction{
		Id:  t.ID[:],
		Raw: t.Raw,
	}
	if t.TxHeader != nil {
		tx.Principal = &pb.AccountId{
			Address: t.Principal.String(),
		}
		tx.Template = &pb.AccountId{
			Address: t.TemplateAddress.String(),
		}
		tx.Method = uint32(t.Method)
		tx.Nonce = &pb.Nonce{
			Counter: t.Nonce,
		}
		tx.Limits = &pb.LayerLimits{
			Min: t.LayerLimits.Min,
			Max: t.LayerLimits.Max,
		}
		tx.MaxGas = t.MaxGas
		tx.GasPrice = t.GasPrice
		tx.MaxSpend = t.MaxSpend
	}
	return tx
}

func convertActivation(a *types.VerifiedActivationTx) *pb.Activation {
	return &pb.Activation{
		Id:        &pb.ActivationId{Id: a.ID().Bytes()},
		Layer:     &pb.LayerNumber{Number: a.PublishEpoch.Uint32()},
		SmesherId: &pb.SmesherId{Id: a.SmesherID.Bytes()},
		Coinbase:  &pb.AccountId{Address: a.Coinbase.String()},
		PrevAtx:   &pb.ActivationId{Id: a.PrevATXID.Bytes()},
		NumUnits:  a.NumUnits,
		Sequence:  a.Sequence,
	}
}
