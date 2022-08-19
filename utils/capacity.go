package utils

import "math"

func CalcEpochCapacity(transactionsNum int64, epochDuration float64, maxTransactionPerSecond uint32) int64 {
	return int64(math.Round(((float64(transactionsNum) / epochDuration) / float64(maxTransactionPerSecond)) * 100.0))
}
