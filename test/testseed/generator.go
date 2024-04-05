package testseed

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strings"
	"time"

	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/signing"

	"github.com/spacemeshos/explorer-backend/model"
	v0 "github.com/spacemeshos/explorer-backend/pkg/transactionparser/v0"
	"github.com/spacemeshos/explorer-backend/storage"
	"github.com/spacemeshos/explorer-backend/utils"
)

// SeedGenerator helper for generate epochs.
type SeedGenerator struct {
	Epochs         SeedEpochs
	Accounts       map[string]AccountContainer
	Activations    map[string]*model.Activation
	Blocks         map[string]*model.Block
	Apps           map[string]model.App
	Layers         map[uint32]*model.Layer
	Rewards        map[string]*model.Reward
	Transactions   map[string]*model.Transaction
	Smeshers       map[string]*model.Smesher
	FirstLayerTime time.Time
	seed           *TestServerSeed
}

// NewSeedGenerator create object which allow fill database for tests.
func NewSeedGenerator(seed *TestServerSeed) *SeedGenerator {
	return &SeedGenerator{
		Epochs:       make(SeedEpochs, 0),
		Accounts:     make(map[string]AccountContainer, 0),
		Apps:         map[string]model.App{},
		seed:         seed,
		Activations:  map[string]*model.Activation{},
		Blocks:       map[string]*model.Block{},
		Layers:       map[uint32]*model.Layer{},
		Rewards:      map[string]*model.Reward{},
		Smeshers:     map[string]*model.Smesher{},
		Transactions: map[string]*model.Transaction{},
	}
}

// GenerateEpoches generate epochs for test.
func (s *SeedGenerator) GenerateEpoches(count int) error {
	now := time.Now()
	result := make([]*SeedEpoch, 0, count)
	var prevEpoch *model.Epoch
	for i := 1; i < count; i++ {
		offset := time.Duration(int64(i)*(int64(s.seed.EpochNumLayers)*int64(s.seed.LayersDuration))) * time.Second
		layerStartDate := now.Add(-1 * offset)
		layersStart := int32(i) * int32(s.seed.EpochNumLayers)
		layersEnd := layersStart + int32(s.seed.EpochNumLayers) - 1
		if i == 1 {
			s.FirstLayerTime = layerStartDate
		}

		seedEpoch := &SeedEpoch{
			Epoch:              s.generateEpoch(int32(i)),
			Layers:             make([]*LayerContainer, 0, s.seed.EpochNumLayers),
			Transactions:       map[string]*model.Transaction{},
			Rewards:            map[string]*model.Reward{},
			Smeshers:           map[string]*model.Smesher{},
			SmeshersCommitment: map[string]int64{},
			Activations:        map[string]*model.Activation{},
			Blocks:             map[string]*model.Block{},
		}

		result = append(result, seedEpoch)

		layerStart := seedEpoch.Epoch.Number * int32(s.seed.EpochNumLayers)
		for j := layerStart; j <= layersEnd; j++ {
			if err := s.fillLayer(j, int32(i), seedEpoch); err != nil {
				return fmt.Errorf("failed to fill layer: %v", err)
			}
			seedEpoch.Epoch.Layers++
		}
		seedEpoch.Epoch.Stats.Current.Decentral = utils.CalcDecentralCoefficient(seedEpoch.SmeshersCommitment)
		duration := float64(s.seed.LayersDuration) * float64(layersEnd-layerStart+1)
		seedEpoch.Epoch.Stats.Current.Capacity = utils.CalcEpochCapacity(seedEpoch.Epoch.Stats.Current.Transactions, duration, uint32(s.seed.MaxTransactionPerSecond))
		if prevEpoch != nil {
			seedEpoch.Epoch.Stats.Cumulative.Capacity = seedEpoch.Epoch.Stats.Current.Capacity
			seedEpoch.Epoch.Stats.Cumulative.Decentral = prevEpoch.Stats.Current.Decentral
			seedEpoch.Epoch.Stats.Cumulative.Smeshers = seedEpoch.Epoch.Stats.Current.Smeshers
			seedEpoch.Epoch.Stats.Cumulative.Transactions = prevEpoch.Stats.Cumulative.Transactions + seedEpoch.Epoch.Stats.Current.Transactions
			seedEpoch.Epoch.Stats.Cumulative.Accounts = seedEpoch.Epoch.Stats.Current.Accounts
			seedEpoch.Epoch.Stats.Cumulative.Rewards = prevEpoch.Stats.Cumulative.Rewards + seedEpoch.Epoch.Stats.Current.Rewards
			seedEpoch.Epoch.Stats.Cumulative.RewardsNumber = prevEpoch.Stats.Cumulative.RewardsNumber + seedEpoch.Epoch.Stats.Current.RewardsNumber
			seedEpoch.Epoch.Stats.Cumulative.Security = prevEpoch.Stats.Current.Security
			seedEpoch.Epoch.Stats.Cumulative.TxsAmount = prevEpoch.Stats.Cumulative.TxsAmount + seedEpoch.Epoch.Stats.Current.TxsAmount

			seedEpoch.Epoch.Stats.Current.Circulation = seedEpoch.Epoch.Stats.Cumulative.Rewards
			seedEpoch.Epoch.Stats.Cumulative.Circulation = seedEpoch.Epoch.Stats.Current.Circulation
		} else {
			seedEpoch.Epoch.Stats.Current.Circulation = seedEpoch.Epoch.Stats.Current.Rewards
			seedEpoch.Epoch.Stats.Cumulative = seedEpoch.Epoch.Stats.Current
		}
		prevEpoch = &seedEpoch.Epoch
	}
	s.Epochs = result
	return nil
}

// SaveEpoches write generated data directly to db.
func (s *SeedGenerator) SaveEpoches(ctx context.Context, db *storage.Storage) error {
	for _, epoch := range s.Epochs {
		if err := db.SaveEpoch(ctx, &epoch.Epoch); err != nil {
			return fmt.Errorf("failed to save epoch: %v", err)
		}
		for _, layerContainer := range epoch.Layers {
			if err := db.SaveLayer(ctx, &layerContainer.Layer); err != nil {
				return fmt.Errorf("failed to save layer: %v", err)
			}
		}
		for _, tx := range epoch.Transactions {
			if err := db.SaveTransaction(ctx, tx); err != nil {
				return fmt.Errorf("failed to save transaction: %v", err)
			}
		}
		for _, reward := range epoch.Rewards {
			if err := db.SaveReward(ctx, reward); err != nil {
				return fmt.Errorf("failed to save reward: %v", err)
			}
		}
		for _, smesher := range epoch.Smeshers {
			if err := db.SaveSmesher(ctx, smesher, uint32(epoch.Epoch.Number)); err != nil {
				return fmt.Errorf("failed to save smesher: %v", err)
			}
		}
		for _, atx := range epoch.Activations {
			if err := db.SaveActivation(ctx, atx); err != nil {
				return fmt.Errorf("failed to save activation: %v", err)
			}
		}
		for _, block := range epoch.Blocks {
			if err := db.SaveBlock(ctx, block); err != nil {
				return fmt.Errorf("failed to save block: %v", err)
			}
		}
	}
	for _, acc := range s.Accounts {
		if err := db.SaveAccount(ctx, acc.layerID, &acc.Account); err != nil {
			return fmt.Errorf("failed to save account: %s", err)
		}
	}
	return nil
}

func (s *SeedGenerator) fillLayer(layerID, epochID int32, seedEpoch *SeedEpoch) error {
	tmpLayer := s.generateLayer(layerID, epochID)
	layerContainer := &LayerContainer{
		Layer:       tmpLayer,
		Blocks:      make([]*BlockContainer, 0),
		Activations: map[string]*model.Activation{},
		Smeshers:    map[string]*model.Smesher{},
	}
	seedEpoch.Layers = append(seedEpoch.Layers, layerContainer)

	for k := 0; k <= rand.Intn(5); k++ {
		tmpAcc, tmpAccSigner := s.generateAccount(tmpLayer.Number)
		s.Accounts[strings.ToLower(tmpAcc.Address)] = AccountContainer{
			layerID:      uint32(layerID),
			Account:      tmpAcc,
			Signer:       tmpAccSigner,
			Transactions: map[string]*model.Transaction{},
			Rewards:      map[string]*model.Reward{},
		}
		tmpApp := model.App{
			Address: tmpAcc.Address,
		}
		s.Apps[tmpApp.Address] = tmpApp

		tmpBl := s.generateBlocks(int32(tmpLayer.Number), seedEpoch.Epoch.Number)

		s.Blocks[tmpBl.Id] = &tmpBl
		blockContainer := &BlockContainer{
			Block:        &tmpBl,
			Transactions: make([]*model.Transaction, 0),
		}
		layerContainer.Blocks = append(layerContainer.Blocks, blockContainer)
		layerContainer.Layer.BlocksNumber++

		for i := 0; i < rand.Intn(3); i++ {
			from := tmpAcc.Address
			to := s.getRandomAcc()
			tmpTx := generateTransaction(i, &tmpLayer, tmpAccSigner, from, to, &tmpBl)

			seedEpoch.Epoch.Stats.Current.Transactions++
			seedEpoch.Epoch.Stats.Current.TxsAmount += int64(tmpTx.Amount)

			layerContainer.Layer.Txs++
			layerContainer.Layer.TxsAmount += tmpTx.Amount
			blockContainer.Block.TxsNumber++
			blockContainer.Block.TxsValue += tmpTx.Amount
			s.saveTransactionForAccount(&tmpTx, from, to)
			seedEpoch.Transactions[tmpTx.Id] = &tmpTx
			s.Transactions[tmpTx.Id] = &tmpTx
			blockContainer.Transactions = append(blockContainer.Transactions, &tmpTx)
		}

		from := tmpAcc.Address
		atxNumUnits := uint32(rand.Intn(1000))
		tmpSm := s.generateSmesher(tmpLayer.Number, from, uint64(atxNumUnits)*s.seed.GetPostUnitsSize())
		layerContainer.Smeshers[tmpSm.Id] = &tmpSm
		seedEpoch.Epoch.Stats.Current.Smeshers++
		seedEpoch.SmeshersCommitment[tmpSm.Id] += int64(tmpSm.CommitmentSize)

		tmpAtx := s.generateActivation(tmpLayer.Number, atxNumUnits, &tmpSm, s.seed.GetPostUnitsSize(), uint32(epochID))
		seedEpoch.Activations[tmpAtx.Id] = &tmpAtx
		layerContainer.Activations[tmpAtx.Id] = &tmpAtx
		seedEpoch.Epoch.Stats.Current.Security += int64(tmpAtx.CommitmentSize)
		s.Activations[tmpAtx.Id] = &tmpAtx

		seedEpoch.Smeshers[strings.ToLower(tmpSm.Id)] = &tmpSm
		blockContainer.SmesherID = tmpSm.Id

		tmpRw := s.generateReward(tmpLayer.Number, &tmpSm)
		seedEpoch.Rewards[tmpRw.Smesher] = &tmpRw
		s.saveReward(&tmpRw)
		seedEpoch.Blocks[tmpBl.Id] = &tmpBl
		s.Rewards[strings.ToLower(tmpRw.Smesher)] = &tmpRw
		s.Smeshers[strings.ToLower(tmpSm.Id)] = &tmpSm
		s.Smeshers[strings.ToLower(tmpSm.Id)].Rewards = int64(tmpRw.Total)
		seedEpoch.Epoch.Stats.Current.RewardsNumber++
		seedEpoch.Epoch.Stats.Current.Rewards += int64(tmpRw.Total)
		seedEpoch.Epoch.Stats.Current.Capacity += int64(tmpSm.CommitmentSize)
	}
	s.Layers[tmpLayer.Number] = &layerContainer.Layer
	return nil
}

func (s *SeedGenerator) getRandomAcc() string {
	for _, val := range s.Accounts {
		return val.Account.Address
	}
	return ""
}

func (s *SeedGenerator) generateActivation(layerNum uint32, atxNumUnits uint32, smesher *model.Smesher, postUnitSize uint64, epoch uint32) model.Activation {
	tx, _ := utils.CalculateLayerStartEndDate(uint32(s.FirstLayerTime.Unix()), layerNum, uint32(s.seed.LayersDuration))
	return model.Activation{
		Id:                strings.ToLower(utils.BytesToHex(randomBytes(32))),
		SmesherId:         smesher.Id,
		Coinbase:          smesher.Coinbase,
		PrevAtx:           strings.ToLower(utils.BytesToHex(randomBytes(32))),
		NumUnits:          atxNumUnits,
		EffectiveNumUnits: atxNumUnits,
		CommitmentSize:    uint64(atxNumUnits) * postUnitSize,
		PublishEpoch:      epoch - 1,
		TargetEpoch:       epoch,
		Received: map[string]int64{
			"collector-test": int64(tx),
		},
	}
}

func (s *SeedGenerator) generateEpoch(epochNum int32) model.Epoch {
	layersStart := uint32(epochNum) * s.seed.EpochNumLayers
	layersEnd := layersStart + s.seed.EpochNumLayers - 1

	epochStart, _ := utils.CalculateLayerStartEndDate(uint32(s.FirstLayerTime.Unix()), layersStart, uint32(s.seed.LayersDuration))
	_, epochEnd := utils.CalculateLayerStartEndDate(uint32(s.FirstLayerTime.Unix()), layersEnd, uint32(s.seed.LayersDuration))
	return model.Epoch{
		Number:     epochNum,
		Start:      epochStart,
		End:        epochEnd,
		LayerStart: layersStart,
		LayerEnd:   layersEnd,
		Layers:     0,
		Stats: model.Stats{
			Current: model.Statistics{
				Capacity:      0,
				Decentral:     0,
				Smeshers:      0,
				Transactions:  0,
				Accounts:      0,
				Circulation:   0,
				Rewards:       0,
				RewardsNumber: 0,
				Security:      0,
				TxsAmount:     0,
			},
			Cumulative: model.Statistics{
				Capacity:      0,
				Decentral:     0,
				Smeshers:      0,
				Transactions:  0,
				Accounts:      0,
				Circulation:   0,
				Rewards:       0,
				RewardsNumber: 0,
				Security:      0,
				TxsAmount:     0,
			},
		},
	}
}

func (s *SeedGenerator) generateLayer(layerNum, epochNum int32) model.Layer {
	start, end := utils.CalculateLayerStartEndDate(uint32(s.FirstLayerTime.Unix()), uint32(layerNum), uint32(s.seed.LayersDuration))
	return model.Layer{
		Number:       uint32(layerNum),
		Status:       2,
		Txs:          0,
		Start:        start,
		End:          end,
		TxsAmount:    0,
		Rewards:      uint64(rand.Intn(1000)),
		Epoch:        uint32(epochNum),
		Hash:         strings.ToLower(fmt.Sprintf("%x", sha256.Sum256(randomBytes(32)))),
		BlocksNumber: 0,
	}
}

func generateTransaction(index int, layer *model.Layer, senderSigner *signing.EdSigner, sender, receiver string, block *model.Block) model.Transaction {
	maxGas := uint64(rand.Intn(1000))
	gasPrice := uint64(rand.Intn(1000))
	return model.Transaction{
		Id:         strings.ToLower(utils.BytesToHex(randomBytes(32))),
		Layer:      layer.Number,
		Block:      block.Id,
		BlockIndex: uint32(index),
		Index:      0,
		State:      int(pb.TransactionState_TRANSACTION_STATE_UNSPECIFIED),
		Timestamp:  layer.Start,
		MaxGas:     maxGas,
		GasPrice:   gasPrice,
		GasUsed:    0,
		Fee:        maxGas * gasPrice,
		Amount:     uint64(rand.Intn(1000)),
		Counter:    uint64(rand.Intn(1000)),
		Type:       3,
		Signature:  strings.ToLower(utils.BytesToHex(randomBytes(30))),
		PublicKey:  senderSigner.PublicKey().String(),
		Sender:     sender,
		Receiver:   receiver,
		SvmData:    "",
	}
}

func (s *SeedGenerator) generateSmesher(layerNum uint32, coinbase string, commitmentSize uint64) model.Smesher {
	tx, _ := utils.CalculateLayerStartEndDate(uint32(s.FirstLayerTime.Unix()), layerNum, uint32(s.seed.LayersDuration))
	return model.Smesher{
		Id:             utils.BytesToHex(randomBytes(32)),
		CommitmentSize: commitmentSize,
		Coinbase:       coinbase,
		AtxCount:       1,
		Timestamp:      uint64(tx),
	}
}

func (s *SeedGenerator) generateReward(layerNum uint32, smesher *model.Smesher) model.Reward {
	tx, _ := utils.CalculateLayerStartEndDate(uint32(s.FirstLayerTime.Unix()), layerNum, uint32(s.seed.LayersDuration))
	return model.Reward{
		Layer:         layerNum,
		Total:         uint64(rand.Intn(1000)),
		LayerReward:   uint64(rand.Intn(1000)),
		LayerComputed: 0,
		Coinbase:      smesher.Coinbase,
		Smesher:       smesher.Id,
		Timestamp:     tx,
	}
}

func (s *SeedGenerator) generateBlocks(layerNum, epochNum int32) model.Block {
	blockStart, blockEnd := utils.CalculateLayerStartEndDate(uint32(s.FirstLayerTime.Unix()), uint32(layerNum), uint32(s.seed.LayersDuration))
	return model.Block{
		Id:        strings.ToLower(utils.NBytesToHex(randomBytes(20), 20)),
		Layer:     uint32(layerNum),
		Epoch:     uint32(epochNum),
		Start:     blockStart,
		End:       blockEnd,
		TxsNumber: 0,
		TxsValue:  0,
	}
}

func (s *SeedGenerator) generateAccount(layerNum uint32) (model.Account, *signing.EdSigner) {
	var key [32]byte
	signer, _ := signing.NewEdSigner()
	copy(key[:], signer.PublicKey().Bytes())
	return model.Account{
		Address: v0.ComputePrincipal(v0.TemplateAddress, &v0.SpawnArguments{
			PublicKey: key,
		}).String(),
		Balance: 0,
		Counter: 0,
		Created: uint64(layerNum),
	}, signer
}

func randomBytes(size int) []byte {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil
	}
	return b
}
