package model

type NetworkInfo struct {
	NetId                    uint32 `json:"netid" bson:"netid"` // nolint will fix it later
	GenesisTime              uint32 `json:"genesis" bson:"genesis"`
	EpochNumLayers           uint32 `json:"layers" bson:"layers"`
	MaxTransactionsPerSecond uint32 `json:"maxtx" bson:"maxtx"`
	LayerDuration            uint32 `json:"duration" bson:"duration"`

	LastLayer          uint32 `json:"lastlayer" bson:"lastlayer"`
	LastLayerTimestamp uint32 `json:"lastlayerts" bson:"lastlayerts"`
	LastApprovedLayer  uint32 `json:"lastapprovedlayer" bson:"lastapprovedlayer"`
	LastConfirmedLayer uint32 `json:"lastconfirmedlayer" bson:"lastconfirmedlayer"`

	ConnectedPeers uint64 `json:"connectedpeers" bson:"connectedpeers"`
	IsSynced       bool   `json:"issynced" bson:"issynced"`
	SyncedLayer    uint32 `json:"syncedlayer" bson:"syncedlayer"`
	TopLayer       uint32 `json:"toplayer" bson:"toplayer"`
	VerifiedLayer  uint32 `json:"verifiedlayer" bson:"verifiedlayer"`
}
