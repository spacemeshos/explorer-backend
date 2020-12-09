package model

type NetworkInfo struct {
    NetId                    uint32
    GenesisTime              uint32
    EpochNumLayers           uint32
    MaxTransactionsPerSecond uint32
    LayerDuration            uint32

    LastLayer                uint32
    LastLayerTimestamp       uint32
    LastApprovedLayer        uint32
    LastConfirmedLayer       uint32

    ConnectedPeers           uint64
    IsSynced                 bool
    SyncedLayer              uint32
    TopLayer                 uint32
    VerifiedLayer            uint32
}
