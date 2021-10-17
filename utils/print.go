package utils

import (
    "github.com/spacemeshos/go-spacemesh/log"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

func PrintLayer(layer *pb.Layer) {
    log.Info("Layer %v, status: %v, blocks: %v, activations: %v",
        layer.Number,
        layer.Status,
        len(layer.Blocks),
        len(layer.Activations),
    )
    for _, atx := range layer.Activations {
        PrintActivation(atx)
    }
    for _, block := range layer.Blocks {
        PrintBlock(block)
    }
}

func PrintBlock(block *pb.Block) {
/*
    log.Info("Block ID: %v, txs: %v",
        block.Id,
        len(block.Transactions),
    )
    for _, tx := range block.Transactions {
        PrintTransaction(tx)
    }
*/
}

func PrintTransaction(tx *pb.Transaction) {
/*
    log.Info("TX ID: %v, sender: %v, gas: %v, amount: %v, counter: %v",
        tx.Id,
        tx.Sender,
        tx.GasOffered,
        tx.Amount,
        tx.Counter,
    )
*/
}

func PrintActivation(atx *pb.Activation) {
    log.Info("ATX ID: %v, layer: %v, smesher: %v, coinbase: %v, prev: %v, size: %v",
        atx.Id,
        atx.Layer,
        atx.SmesherId,
        atx.Coinbase,
        atx.PrevAtx,
        atx.NumUnits,
    )
}

func PrintReward(reward *pb.Reward) {
/*
    log.Info("Reward layer: %v, total: %v, layerReward: %v, computed: %v, coinbase: %v, smesher: %v",
        reward.Layer,
        reward.Total,
        reward.LayerReward,
        reward.LayerComputed,
        reward.Coinbase,
        reward.Smesher,
    )
*/
}
