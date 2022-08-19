package testseed

import (
	"github.com/spacemeshos/explorer-backend/model"
)

// AccountContainer ...
type AccountContainer struct {
	layerID      uint32
	Account      model.Account
	Transactions map[string]*model.Transaction
	Rewards      map[string]*model.Reward
}

func (s *SeedGenerator) saveTransactionForAccount(tx *model.Transaction, accountFrom, accountTo string) {
	s.Accounts[accountFrom].Transactions[tx.Id] = tx
	s.Accounts[accountTo].Transactions[tx.Id] = tx
}

func (s *SeedGenerator) saveReward(reward *model.Reward) {
	s.Accounts[reward.Coinbase].Rewards[reward.Smesher] = reward
}
