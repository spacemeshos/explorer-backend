package testseed

import (
	"strings"

	"github.com/spacemeshos/explorer-backend/model"
)

// AccountContainer is a container for accounts with transactions and rewards belongs to generated account.
type AccountContainer struct {
	layerID      uint32
	Account      model.Account
	Transactions map[string]*model.Transaction
	Rewards      map[string]*model.Reward
}

func (s *SeedGenerator) saveTransactionForAccount(tx *model.Transaction, accountFrom, accountTo string) {
	s.Accounts[strings.ToLower(accountFrom)].Transactions[tx.Id] = tx
	s.Accounts[strings.ToLower(accountTo)].Transactions[tx.Id] = tx

	tmpContainerFrom := s.Accounts[strings.ToLower(accountFrom)]
	tmpContainerFrom.Account.Txs++
	s.Accounts[strings.ToLower(accountFrom)] = tmpContainerFrom

	if accountTo == accountFrom {
		return
	}

	tmpContainerTo := s.Accounts[strings.ToLower(accountTo)]
	tmpContainerTo.Account.Txs++
	s.Accounts[strings.ToLower(accountTo)] = tmpContainerTo
}

func (s *SeedGenerator) saveReward(reward *model.Reward) {
	s.Accounts[strings.ToLower(reward.Coinbase)].Rewards[reward.Smesher] = reward
}
