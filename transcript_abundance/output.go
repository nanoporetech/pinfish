package main

import (
	"fmt"
	"sort"
)

type TrsAbd []CompRecord

func (s TrsAbd) Len() int {
	return len(s)
}
func (s TrsAbd) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s TrsAbd) Less(i, j int) bool {
	return s[i].Prob > s[j].Prob
}

func SaveCounts(abd map[string]float64, totalReads int) {
	fmt.Println("Reference\tCount")
	trsAbd := make(TrsAbd, 0, len(abd))

	for tr, p := range abd {
		trsAbd = append(trsAbd, CompRecord{tr, p * float64(totalReads)})
	}

	sort.Sort(trsAbd)

	for _, tr := range trsAbd {
		fmt.Printf("%s\t%f\n", tr.Target, tr.Prob)
	}
}
