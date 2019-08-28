package main

import (
	"github.com/biogo/biogo/feat/gene"
	"sort"
	"strings"
)

// Sort transcripts by chromosome and coordinates.
func SortTranscripts(trs []*gene.CodingTranscript) []*gene.CodingTranscript {
	sort.Sort(byCoord(trs))
	return trs
}

type byCoord []*gene.CodingTranscript

func (s byCoord) Len() int {
	return len(s)
}

func (s byCoord) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byCoord) Less(i, j int) bool {
	if s[i].Location().Name() != s[j].Location().Name() {
		if strings.Compare(s[i].Location().Name(), s[j].Location().Name()) == -1 {
			return true
		}
	} else {
		si := s[i].Location().Start() + s[i].Start()
		sj := s[j].Location().Start() + s[j].Start()
		if si != sj {
			return si < sj
		} else {
			li := s[i].Len()
			lj := s[j].Len()
			return li < lj
		}
	}
	return false
}
