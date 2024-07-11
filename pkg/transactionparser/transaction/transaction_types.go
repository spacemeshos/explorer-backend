package transaction

const (
	// TypeSpawn is type of the spawn transaction.
	TypeSpawn = 1 + iota
	// TypeMultisigSpawn is type of the multisig spawn transaction.
	TypeMultisigSpawn
	// TypeSpend is type of the spend transaction.
	TypeSpend
	// TypeMultisigSpend is type of the multisig spend transaction.
	TypeMultisigSpend
	TypeVestingSpawn
	TypeVaultSpawn
	TypeDrainVault
)
