package main

import (
	"fmt"
	"gopkg.in/cheggaaa/pb.v2"
	"os"
)

// Compatibility structure:
type CompRecord struct {
	Target string
	Prob   float64
}

// Master structure:
type TranscriptPool struct {
	Compat         map[string][]CompRecord
	MinReadLength  int
	FullLenMax     int
	ScoreThreshold float64
	AlnThreshold   float64
}

// Create new master structure:
func NewTranscriptPool(minReadLength int, scoreThreshold float64, alnThreshold float64, fullLenMax int) *TranscriptPool {
	p := new(TranscriptPool)
	p.Compat = make(map[string][]CompRecord)
	p.MinReadLength = minReadLength
	p.FullLenMax = fullLenMax
	p.ScoreThreshold = scoreThreshold
	p.AlnThreshold = alnThreshold

	if VERBOSE {
		L.Printf("Minimum read length is %d.\n", p.MinReadLength)
		L.Printf("Maximum distance of full length read from start is %d.\n", p.FullLenMax)
		L.Printf("Score threshold is %f.\n", p.ScoreThreshold)
		L.Printf("Alignment threshold is %f.\n", p.AlnThreshold)
	}

	return p
}

// Load compatibility information from reads:
func (p *TranscriptPool) LoadCompatibility(pafChan chan *PafRecord) {
	prevReadName := ""
	var currRecords []*PafRecord

	recCount := 0
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
		recCount++
	}
	// Process loast read:
	if len(currRecords) > 0 {
		p.GetCompatibility(currRecords)
	}

	if VERBOSE {
		L.Printf("Loaded %d alignment records for %d reads.\n", recCount, len(p.Compat))
	}

}

func (p *TranscriptPool) GetCompatibility(recs []*PafRecord) {

	// Minimum distance from start to be considered full length:
	// FIXME:
	fullLengthMinDistance := p.FullLenMax
	// Minimum read length:
	minReadLength := p.MinReadLength
	// Minimum score threshold when considering equivalence:
	threshold := p.ScoreThreshold
	// Minimum fraction aligned of best match:
	alignmentThreshold := p.AlnThreshold

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
	var bar *pb.ProgressBar
	if VERBOSE {
		L.Printf("Estimating transcript abundances by EM:\n")
		bar = pb.StartNew(niter)
	}

	for i := 0; i < niter; i++ {
		// Estimate abundances from compatibilities:
		abundances := p.Abundances()
		// Update compatibilities:
		p.UpdateCompatibility(abundances)
		if VERBOSE {
			bar.Add(1)
		}

	}

	if VERBOSE {
		bar.Finish()
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

// Save compatibility information:
func (p *TranscriptPool) SaveCompatibilities(cf string) {
	fh, err := os.Create(cf)
	if err != nil {
		L.Fatalf("Could not create compatibility file %s: %s\n", cf, err)
	}
	if VERBOSE {
		L.Printf("Saving compatibility information to %s.\n", cf)
	}

	for read, comps := range p.Compat {
		line := read + "\t"
		for i, c := range comps {
			line += c.Target
			if i != len(comps)-1 {
				line += ","
			}
		}
		line += "\t"
		for i, c := range comps {
			line += fmt.Sprintf("%f", c.Prob)
			if i != len(comps)-1 {
				line += ","
			}
		}
		line += "\n"
		fh.WriteString(line)
	}
	fh.Close()
}
