package main

import (
	"github.com/biogo/biogo/feat"
	"github.com/biogo/biogo/feat/gene"
	"github.com/google/uuid"
	"math"
)

// Structure to hold a 3' locus:
type Locus struct {
	Chrom        string
	Orient       feat.Orientation
	ThreePrime   float64
	SampleWeight float64
	Id           string
}

// Map holding transcripts at each 3' locus:
type LocusPool map[Locus][]*gene.CodingTranscript

// Get the 3' end of a transcript:
func GetThreePrime(tr *gene.CodingTranscript) int {
	switch tr.Orient {
	case feat.Forward:
		return tr.Location().Start() + tr.End()
	case feat.Reverse:
		return tr.Location().Start() + tr.Start()
	case feat.NotOriented: // As a convention pick the start position.
		return tr.Location().Start() + tr.Start()
	}
	return -1

}

// Look for a compatible 3' locus in the pool based on the distance from the locus mean position.
func SearchLoci(tr *gene.CodingTranscript, pool LocusPool, threeTol int) (Locus, bool) {
	// Empty pool, nothing found:
	if len(pool) == 0 {
		return Locus{}, false
	}

	// Unoriented transcripts ar all registered
	// as individual 3' loci:
	if tr.Orient == feat.NotOriented {
		return Locus{}, false
	}

	// Check each locus for compatibility:
	for locus, _ := range pool {
		// Chromosome mistmatch:
		if locus.Chrom != tr.Location().Name() {
			continue
		}
		// Orientation mismatch:
		if locus.Orient != tr.Orient {
			continue
		}

		// Calculate the distance of transcript 3' from locus mean position:
		delta := math.Abs(float64(GetThreePrime(tr)) - locus.ThreePrime)

		// Apply 3' tolerance:
		if delta < float64(threeTol) {
			return locus, true
		}

	}
	return Locus{}, false
}

// Load transcriopt into loci defined by the 3' ends.
func LoadLoci(trsChan chan *gene.CodingTranscript, threeTol int, monoDiscard bool, unorientDiscard bool) LocusPool {
	locusPool := make(LocusPool)

	for tr := range trsChan {
		// Discard monoexonic transcripts if requested:
		if monoDiscard && len(tr.Exons()) == 1 {
			continue
		}
		// Discard transcripts with no orientation if requested:
		if unorientDiscard && (tr.Orient == feat.NotOriented) {
			continue
		}

		// Search for the first compatible locus:
		locus, found := SearchLoci(tr, locusPool, threeTol)

		if !found {
			// New locus:
			newLocus := Locus{tr.Location().Name(), tr.Orient, float64(GetThreePrime(tr)), 1, uuid.New().String()}
			// Set locus id in description string:
			tr.Desc = "\"" + newLocus.Id + "\"" + "\n" + tr.Desc
			// Add to locus pool:
			locusPool[newLocus] = []*gene.CodingTranscript{tr}
		} else {
			// Set locus id in description string:
			tr.Desc = "\"" + locus.Id + "\"" + "\n" + tr.Desc
			// Add transcript to locus buffer:
			trs := append(locusPool[locus], tr)
			// Remove locus:
			delete(locusPool, locus)
			// Calculate the new mean 3' end of the locus:
			locus.ThreePrime = (locus.ThreePrime*locus.SampleWeight + float64(GetThreePrime(tr))) / (locus.SampleWeight + 1)
			locus.SampleWeight++ // Update locus size.
			// Add updated locus to pool:
			locusPool[locus] = trs
		}
	}

	return locusPool
}
