package utils

import (
    "sort"
)

type CommitmentSizes []int64

func (a CommitmentSizes) Len() int           { return len(a) }
func (a CommitmentSizes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CommitmentSizes) Less(i, j int) bool { return a[i] < a[j] }
/*
        1          sum((n + 1 - i)*y[i])
    G = -(n + 1 - 2---------------------
        n                sum(y[i])
*/
func Gini(smeshers map[string]int64) float64 {
    var n int
    var sum float64
    if len(smeshers) > 0 {
        data := make(CommitmentSizes, len(smeshers))
        for _, commitment_size := range smeshers {
            data[n] = commitment_size
            if data[n] == 0 {
                data[n] = 1
            }
            sum += float64(data[n])
            n++
        }
        if sum > 0 && n > 0 {
            sort.Sort(data)
            var top float64
            for i, y := range data {
                top += float64(int64(n - i) * y)
            }
            c := (float64(n) + 1.0 - 2.0 * top / sum) / float64(n)
//            log.Info("gini: top = %v, n = %v, sum = %v, coef = %v", top, n, sum, c)
            return c
        }
    }
    return 1
}

