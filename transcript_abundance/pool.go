package main

import (
//"fmt"
)

type CompRecord struct {
	Target string
	Prob   float64
}

type TranscriptPool struct {
	Compat map[string][]CompRecord
}

func NewTranscriptPool() *TranscriptPool {
	p := new(TranscriptPool)
	p.Compat = make(map[string][]CompRecord)
	return p
}

func (p *TranscriptPool) LoadCompatibility(pafChan chan *PafRecord) {
	prevReadName := ""
	var currRecords []*PafRecord

	for pafRec := range pafChan {
		if pafRec.QueryName != prevReadName && len(currRecords) > 0 {
			//fmt.Println(len(currRecords), prevReadName)
			p.GetCompatibility(currRecords)
			currRecords = make([]*PafRecord, 0, 1000)
		}
		prevReadName = pafRec.QueryName
		currRecords = append(currRecords, pafRec)
	}
	p.GetCompatibility(currRecords)

}

func (p *TranscriptPool) GetCompatibility(recs []*PafRecord) {

	fullLengthMinDistance := 20
	minReadLength := 0
	threshold := 0.95
	alignmentThreshold := 0.5

	//Determine best match:
	readLength := recs[0].QueryLength
	bestMatchAlignLen := 0
	bestNumMatches := 0
	bestIsFullLength := false

	for _, rec := range recs {
		fl := rec.TargetStart < fullLengthMinDistance
		if rec.NrMatches > bestNumMatches || (rec.NrMatches == bestNumMatches && fl) {
			bestMatchAlignLen = rec.AlnLength
			bestNumMatches = rec.NrMatches
			bestIsFullLength = fl
		}
	}

	// Skip read if fails filters:
	fractionAligned := float64(bestMatchAlignLen) / float64(readLength)
	if fractionAligned < alignmentThreshold || readLength < minReadLength {
		return // FIXME
	}

	// Is equivalent hit if score is high enough
	// an full length status matches:
	isEquHit := func(r *PafRecord) bool {
		f := float64(r.NrMatches) / float64(bestNumMatches)
		l := r.TargetStart < fullLengthMinDistance
		return (f > threshold) && (l == bestIsFullLength)
	}

	// Count equivalent hits:
	numHits := 0
	for _, rec := range recs {
		if isEquHit(rec) {
			numHits++
		}
	}

	// "Distribute" the hits equally between the targets:
	for _, rec := range recs {
		if isEquHit(rec) {
			p.Compat[rec.QueryName] = append(p.Compat[rec.QueryName], CompRecord{rec.TargetName, 1.0 / float64(numHits)})
		}
	}
}

func (p *TranscriptPool) EmEstimate(niter int) {
	for i := 0; i < niter; i++ {
		abundances := p.Abundances()
		p.UpdateCompatibility(abundances)
	}
}

func (p *TranscriptPool) Abundances() map[string]float64 {
	res := make(map[string]float64)
	total := 0.0

	for _, hits := range p.Compat {
		for _, h := range hits {
			res[h.Target] += h.Prob
			total += h.Prob
		}
	}

	for tr, ab := range res {
		res[tr] = ab / total
	}

	return res
}

func (p *TranscriptPool) UpdateCompatibility(abundances map[string]float64) {

	for read, hits := range p.Compat {
		ids := make([]string, 0, len(hits))

		total := 0.0
		for _, h := range hits {
			total += abundances[h.Target]
			ids = append(ids, h.Target)
		}

		p.Compat[read] = nil
		for _, id := range ids {
			p.Compat[read] = append(p.Compat[read], CompRecord{id, abundances[id] / total})
		}
	}

}
