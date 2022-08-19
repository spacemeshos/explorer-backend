package utils

// CalculateLayerStartEndDate ...
func CalculateLayerStartEndDate(genesisTime, layerNum, layerDuration uint32) (layerStartDate, layerEndDate uint32) {
	if layerNum == 0 {
		layerStartDate = genesisTime
	} else {
		layerStartDate = genesisTime + (layerNum-1)*layerDuration
	}
	layerEndDate = layerStartDate + layerDuration - 1
	return layerStartDate, layerEndDate
}
