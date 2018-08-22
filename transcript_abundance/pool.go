package main

import (
//"fmt"
)

// Compatibility structure:
type CompRecord struct {
	Target string
	Prob   float64
}

// Master structure:
type TranscriptPool struct {
	Compat map[string][]CompRecord
}

// Create new master structure:
func NewTranscriptPool() *TranscriptPool {
	p := new(TranscriptPool)
	p.Compat = make(map[string][]CompRecord)
	return p
}

// Load compatibility information from reads:
func (p *TranscriptPool) LoadCompatibility(pafChan chan *PafRecord) {
	prevReadName := ""
	var currRecords []*PafRecord

	// Iterate over PAF records:
	for pafRec := range pafChan {
		// If read changes...
		if pafRec.QueryName != prevReadName && len(currRecords) > 0 {
			// Store compatibilities:
			p.GetCompatibility(currRecords)
			// Reste buffer of current records:
			currRecords = make([]*PafRecord, 0, 1000)
		}
		prevReadName = pafRec.QueryName
		// Add PAF record to buffer:

		currRecords = append(currRecords, pafRec)
	}
	// Process loast read:
	p.GetCompatibility(currRecords)

}

func (p *TranscriptPool) GetCompatibility(recs []*PafRecord) {

	// Minimum distance from start to be considered full length:
	// FIXME:
	fullLengthMinDistance := 20
	// Minimum read length:
	minReadLength := 0
	// Minimum score threshold when considering equivalence:
	threshold := 0.95
	// Minimum fraction aligned of best match:
	alignmentThreshold := 0.5

	//Determine best match:
	readLength := recs[0].QueryLength
	bestMatchAlignLen := 0
	bestNumMatches := 0
	bestIsFullLength := false

	// Iterate over buffer:
	for _, rec := range recs {
		// Is read full length:
		fl := rec.TargetStart < fullLengthMinDistance
		// Found better score or euqivalent full length match:
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
	// and full length status matches:
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

// Estimate abundances by EM:
func (p *TranscriptPool) EmEstimate(niter int) {
	for i := 0; i < niter; i++ {
		// Estimate abundances from compatibilities:
		abundances := p.Abundances()
		// Update compatibilities:
		p.UpdateCompatibility(abundances)
	}
}

// Calculate transcript abundances:
func (p *TranscriptPool) Abundances() map[string]float64 {
	res := make(map[string]float64)
	total := 0.0

	// Sum up hits for each transcript:
	for _, hits := range p.Compat {
		for _, h := range hits {
			res[h.Target] += h.Prob
			total += h.Prob
		}
	}

	// Normalise abundances by total hits:
	for tr, ab := range res {
		res[tr] = ab / total
	}

	return res
}

// Update compatibilities:
func (p *TranscriptPool) UpdateCompatibility(abundances map[string]float64) {

	for read, hits := range p.Compat {
		ids := make([]string, 0, len(hits))

		// Figure out total abundance for
		// each read:
		total := 0.0
		for _, h := range hits {
			total += abundances[h.Target]
			ids = append(ids, h.Target)
		}

		// Update compatibility - abundance of belonging transcript normalised by total:
		p.Compat[read] = nil
		for _, id := range ids {
			p.Compat[read] = append(p.Compat[read], CompRecord{id, abundances[id] / total})
		}
	}

}
