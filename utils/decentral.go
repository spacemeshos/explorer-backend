package utils

import (
	"math"
)

func CalcDecentralCoefficient(smeshers map[string]int64) int64 {
	a := math.Min(float64(len(smeshers)), 1e4)
	return int64(100.0 * (0.5*(a*a)/1e8 + 0.5*(1.0-Gini(smeshers))))
}
