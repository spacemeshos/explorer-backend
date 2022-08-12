package testseed

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"time"

	"github.com/spacemeshos/explorer-backend/model"
	"github.com/spacemeshos/explorer-backend/storage"
	"github.com/spacemeshos/explorer-backend/utils"
)

// SeedGenerator test helper for generate epochs.
type SeedGenerator struct {
	storage  *storage.Storage
	Epochs   SeedEpochs
	Accounts map[string]AccountContainer
	Apps     map[string]model.App
}

// NewSeedGenerator create object which allow fill database for tests.
func NewSeedGenerator(db *storage.Storage) *SeedGenerator {
	return &SeedGenerator{
		storage:  db,
		Epochs:   make(SeedEpochs, 0),
		Accounts: make(map[string]AccountContainer, 0),
		Apps:     map[string]model.App{},
	}
}

// GenerateEpoches ...
func (s *SeedGenerator) GenerateEpoches(count int) error {
	now := time.Now()
	seed := GetServerSeed()
	result := make([]*SeedEpoch, 0, count)
	for i := 1; i < count; i++ {
		offset := time.Duration(int64(i)*int64(seed.EpochNumLayers*seed.LayersDuration)) * time.Second
		layerDuration := time.Duration(seed.EpochNumLayers*seed.LayersDuration) * time.Second
		layerStartDate := now.Add(-1 * offset)
		layerEndDate := layerStartDate.Add(layerDuration)
		layersStart := int32(i) * int32(seed.EpochNumLayers)
		layersEnd := layersStart + int32(seed.EpochNumLayers) - 1

		seedEpoch := &SeedEpoch{
			Epoch:        generateEpoch(int32(i), layerStartDate, layerEndDate, layersStart, layersEnd),
			Layers:       make([]model.Layer, 0, seed.EpochNumLayers),
			Transactions: map[string]model.Transaction{},
			Rewards:      map[string]model.Reward{},
			Smeshers:     map[string]model.Smesher{},
			Activations:  map[string]model.Activation{},
			Blocks:       map[string]model.Block{},
		}
		if err := s.storage.SaveEpoch(context.TODO(), &seedEpoch.Epoch); err != nil {
			return fmt.Errorf("failed to save epoch: %v", err)
		}
		result = append(result, seedEpoch)

		layerStart := seedEpoch.Epoch.Number * int32(seed.EpochNumLayers)
		for j := layerStart; j < layersEnd; j++ {
			if err := s.fillLayer(j, int32(i), seedEpoch); err != nil {
				return fmt.Errorf("failed to fill layer: %v", err)
			}
		}
	}
	s.Epochs = result
	return nil
}

func (s *SeedGenerator) fillLayer(layerID, epochID int32, seedEpoch *SeedEpoch) error {
	tmpLayer := generateLayer(layerID, epochID)
	if err := s.storage.SaveLayer(context.TODO(), &tmpLayer); err != nil {
		return fmt.Errorf("failed to save layer: %v", err)
	}
	seedEpoch.Layers = append(seedEpoch.Layers, tmpLayer)
	for k := 0; k <= rand.Intn(3); k++ {
		tmpAcc := generateAccount()
		if err := s.storage.SaveAccount(context.TODO(), uint32(layerID), &tmpAcc); err != nil {
			return fmt.Errorf("failed to save account: %s", err)
		}
		s.Accounts[tmpAcc.Address] = AccountContainer{
			Account:      tmpAcc,
			Transactions: map[string]*model.Transaction{},
			Rewards:      map[string]*model.Reward{},
		}
		tmpApp := model.App{
			Address: tmpAcc.Address,
		}
		if err := s.storage.SaveApp(context.TODO(), &tmpApp); err != nil {
			return fmt.Errorf("failed to save app: %s", err)
		}
		s.Apps[tmpApp.Address] = tmpApp
	}
	for k := 0; k < rand.Intn(3); k++ {
		tmpBl := generateBlocks(int32(tmpLayer.Number), seedEpoch.Epoch.Number)
		if err := s.storage.SaveBlock(context.TODO(), &tmpBl); err != nil {
			return fmt.Errorf("failed to save block: %v", err)
		}
		seedEpoch.Blocks[tmpBl.Id] = tmpBl
	}

	for k := 0; k < rand.Intn(3); k++ {
		from := s.getRandomAcc()
		to := s.getRandomAcc()
		tmpTx := generateTransaction(tmpLayer.Number, from, to)
		if err := s.storage.SaveTransaction(context.TODO(), &tmpTx); err != nil {
			return fmt.Errorf("failed to save transaction: %v", err)
		}
		s.saveTransactionForAccount(&tmpTx, from, to)
		seedEpoch.Transactions[tmpTx.Id] = tmpTx
	}

	for k := 0; k < rand.Intn(3); k++ {
		tmpAtx := generateActivation(tmpLayer.Number)
		if err := s.storage.SaveActivation(context.TODO(), &tmpAtx); err != nil {
			return fmt.Errorf("failed to save activation: %v", err)
		}
		seedEpoch.Activations[tmpAtx.Id] = tmpAtx

		from := s.getRandomAcc()

		tmpSm := generateSmesher(tmpAtx.SmesherId, from)
		if err := s.storage.SaveSmesher(context.TODO(), &tmpSm); err != nil {
			return fmt.Errorf("failed to save smesher: %v", err)
		}
		seedEpoch.Smeshers[tmpSm.Id] = tmpSm

		tmpRw := generateReward(tmpLayer.Number, tmpSm.Id, from)
		if err := s.storage.SaveReward(context.TODO(), &tmpRw); err != nil {
			return fmt.Errorf("failed to save reward: %v", err)
		}
		seedEpoch.Rewards[tmpRw.Smesher] = tmpRw
		s.saveReward(&tmpRw, from)
	}
	return nil
}

func (s *SeedGenerator) getRandomAcc() string {
	for _, val := range s.Accounts {
		return val.Account.Address
	}
	return ""
}

func generateActivation(layerNum uint32) model.Activation {
	return model.Activation{
		Id:        hashFromRandomBytes(),
		Layer:     layerNum,
		SmesherId: hashFromRandomBytes(),
		Coinbase:  hashFromRandomBytes(),
		PrevAtx:   hashFromRandomBytes(),
		NumUnits:  0,
		Timestamp: 0,
	}
}

func generateEpoch(epochNum int32, from, to time.Time, layersStart, layersEnd int32) model.Epoch {
	return model.Epoch{
		Number:     epochNum,
		Start:      uint32(from.Unix()),
		End:        uint32(to.Unix()),
		LayerStart: uint32(layersStart),
		LayerEnd:   uint32(layersEnd),
		Layers:     uint32(rand.Intn(1000)),
		Stats: model.Stats{
			Current: model.Statistics{
				Capacity:      int64(rand.Intn(1000)),
				Decentral:     int64(rand.Intn(1000)),
				Smeshers:      int64(rand.Intn(1000)),
				Transactions:  int64(rand.Intn(1000)),
				Accounts:      int64(rand.Intn(1000)),
				Circulation:   int64(rand.Intn(1000)),
				Rewards:       int64(rand.Intn(1000)),
				RewardsNumber: int64(rand.Intn(1000)),
				Security:      int64(rand.Intn(1000)),
				TxsAmount:     int64(rand.Intn(1000)),
			},
			Cumulative: model.Statistics{
				Capacity:      int64(rand.Intn(1000)),
				Decentral:     int64(rand.Intn(1000)),
				Smeshers:      int64(rand.Intn(1000)),
				Transactions:  int64(rand.Intn(1000)),
				Accounts:      int64(rand.Intn(1000)),
				Circulation:   int64(rand.Intn(1000)),
				Rewards:       int64(rand.Intn(1000)),
				RewardsNumber: int64(rand.Intn(1000)),
				Security:      int64(rand.Intn(1000)),
				TxsAmount:     int64(rand.Intn(1000)),
			},
		},
	}
}

func generateLayer(layerNum, epochNum int32) model.Layer {
	return model.Layer{
		Number:       uint32(layerNum),
		Status:       2,
		Txs:          0,
		Start:        1660227073,
		End:          1660227173,
		TxsAmount:    uint64(rand.Intn(1000)),
		AtxNumUnits:  uint64(rand.Intn(1000)),
		Rewards:      uint64(rand.Intn(1000)),
		Epoch:        uint32(epochNum),
		Smeshers:     uint32(rand.Intn(1000)),
		Hash:         hashFromRandomBytes(),
		BlocksNumber: uint32(rand.Intn(1000)),
	}
}

func generateTransaction(layerNum uint32, sender, receiver string) model.Transaction {
	return model.Transaction{
		Id:          utils.BytesToHex(randomBytes(32)),
		Layer:       layerNum,
		Block:       "",
		BlockIndex:  0,
		Index:       0,
		State:       0,
		Timestamp:   0,
		GasProvided: 0,
		GasPrice:    0,
		GasUsed:     0,
		Fee:         uint64(rand.Intn(1000)),
		Amount:      uint64(rand.Intn(1000)),
		Counter:     0,
		Type:        0,
		Scheme:      0,
		Signature:   hashFromRandomBytes(),
		PublicKey:   hashFromRandomBytes(),
		Sender:      sender,
		Receiver:    receiver,
		SvmData:     hashFromRandomBytes(),
	}
}

func generateSmesher(smesherID, coinbaseAddr string) model.Smesher {
	return model.Smesher{
		Id:             smesherID,
		Geo:            model.Geo{},
		CommitmentSize: 0,
		Coinbase:       coinbaseAddr,
		AtxCount:       0,
		Timestamp:      0,
	}
}

func generateReward(layerNum uint32, smesherID, coinbaseAddr string) model.Reward {
	return model.Reward{
		Layer:         layerNum,
		Total:         uint64(rand.Intn(1000)),
		LayerReward:   uint64(rand.Intn(1000)),
		LayerComputed: 0,
		Coinbase:      coinbaseAddr,
		Smesher:       smesherID,
		Space:         0,
		Timestamp:     uint32(time.Now().Unix()),
	}
}

func generateBlocks(layerNum, epochNum int32) model.Block {
	return model.Block{
		Id:        hashFromRandomBytes(),
		Layer:     uint32(layerNum),
		Epoch:     uint32(epochNum),
		Start:     uint32(rand.Intn(1000)),
		End:       uint32(rand.Intn(1000)),
		TxsNumber: uint32(rand.Intn(1000)),
		TxsValue:  uint64(rand.Intn(1000)),
	}
}

func generateAccount() model.Account {
	return model.Account{
		Address: utils.BytesToAddressString(randomBytes(20)),
		Balance: 0,
		Counter: 0,
	}
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

func hashFromRandomBytes() string {
	return fmt.Sprintf("%x", sha256.Sum256(randomBytes(32)))
}
