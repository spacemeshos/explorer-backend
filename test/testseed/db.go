package testseed

import (
	"errors"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/explorer-backend/utils"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/genvm/sdk"
	sdkWallet "github.com/spacemeshos/go-spacemesh/genvm/sdk/wallet"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"
	"github.com/spacemeshos/go-spacemesh/sql"
	"strings"
)

type Client struct {
	SeedGen *SeedGenerator
}

const (
	methodSend = 16
)

func (c *Client) GetLayer(db *sql.Database, lid types.LayerID, numLayers uint32) (*pb.Layer, error) {
	for _, epoch := range c.SeedGen.Epochs {
		for _, layerContainer := range epoch.Layers {
			if layerContainer.Layer.Number != lid.Uint32() {
				continue
			}
			atx := make([]*pb.Activation, 0, len(layerContainer.Activations))
			for _, atxGenerated := range layerContainer.Activations {
				smesherId, _ := utils.StringToBytes(atxGenerated.SmesherId)
				atx = append(atx, &pb.Activation{
					Id:        &pb.ActivationId{Id: mustParse(atxGenerated.Id)},
					Layer:     &pb.LayerNumber{Number: atxGenerated.Layer},
					SmesherId: &pb.SmesherId{Id: smesherId},
					Coinbase:  &pb.AccountId{Address: atxGenerated.Coinbase},
					PrevAtx:   &pb.ActivationId{Id: mustParse(atxGenerated.PrevAtx)},
					NumUnits:  atxGenerated.NumUnits,
				})
			}

			blocksRes := make([]*pb.Block, 0)
			for _, blockContainer := range layerContainer.Blocks {
				tx := make([]*pb.Transaction, 0, len(blockContainer.Transactions))
				for _, txContainer := range blockContainer.Transactions {
					receiver, err := types.StringToAddress(txContainer.Receiver)
					if err != nil {
						panic("invalid receiver address: " + err.Error())
					}
					signer := c.SeedGen.Accounts[strings.ToLower(txContainer.Sender)].Signer
					tx = append(tx, &pb.Transaction{
						Id:     mustParse(txContainer.Id),
						Method: methodSend,
						Principal: &pb.AccountId{
							Address: txContainer.Sender,
						},
						GasPrice: txContainer.GasPrice,
						MaxGas:   txContainer.MaxGas,
						Nonce: &pb.Nonce{
							Counter: txContainer.Counter,
						},
						Template: &pb.AccountId{
							Address: wallet.TemplateAddress.String(),
						},
						Raw: sdkWallet.Spend(signer.PrivateKey(), receiver, txContainer.Amount, types.Nonce(txContainer.Counter), sdk.WithGasPrice(txContainer.GasPrice)),
					})
				}
				smesherId, _ := utils.StringToBytes(blockContainer.SmesherID)
				blocksRes = append(blocksRes, &pb.Block{
					Id:           mustParse(blockContainer.Block.Id),
					Transactions: tx,
					SmesherId: &pb.SmesherId{
						Id: smesherId,
					},
				})
			}
			pbLayer := &pb.Layer{
				Number:      &pb.LayerNumber{Number: layerContainer.Layer.Number},
				Status:      pb.Layer_LayerStatus(layerContainer.Layer.Status),
				Hash:        mustParse(layerContainer.Layer.Hash),
				Blocks:      blocksRes,
				Activations: atx,
			}

			return pbLayer, nil
		}
	}

	return nil, errors.New("could not find layer")
}

func (c *Client) GetLayerRewards(db *sql.Database, lid types.LayerID) (rst []*types.Reward, err error) {
	for _, epoch := range c.SeedGen.Epochs {
		for _, reward := range epoch.Rewards {
			if reward.Layer != lid.Uint32() {
				continue
			}

			coinbase, _ := utils.StringToBytes(reward.Coinbase)
			var addr types.Address
			copy(addr[:], coinbase)

			r := &types.Reward{
				Layer:       types.LayerID(reward.Layer),
				TotalReward: reward.Total,
				LayerReward: reward.LayerReward,
				Coinbase:    addr,
			}

			rst = append(rst, r)
		}
	}

	return rst, nil
}

func (c *Client) AccountsSnapshot(db *sql.Database, lid types.LayerID) (rst []*types.Account, err error) {
	for _, accountContainer := range c.SeedGen.Accounts {
		if accountContainer.layerID != lid.Uint32() {
			continue
		}

		accAddr, _ := utils.StringToBytes(accountContainer.Account.Address)
		var addr types.Address
		copy(addr[:], accAddr)

		rst = append(rst, &types.Account{
			Layer:   types.LayerID(accountContainer.layerID),
			Address: addr,
			Balance: accountContainer.Account.Balance,
		})
	}

	return rst, nil
}

func mustParse(str string) []byte {
	res, err := utils.StringToBytes(str)
	if err != nil {
		panic("error while parse string to bytes: " + err.Error())
	}
	return res
}
