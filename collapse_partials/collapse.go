package main

import (
	"fmt"
	"github.com/biogo/biogo/feat"
	"github.com/biogo/biogo/feat/gene"
	"strconv"
	"strings"
)

// Parse out locus name and cluster size from transcriot description.
func parseDesc(desc string) (string, int) {
	descTmp := strings.Split(desc, "\n")
	desc, clStr := descTmp[0], descTmp[1]
	clSize, _ := strconv.Atoi(clStr)
	return desc, clSize
}

// "Merge" the second transcript into the first one by assigning it the sum of
// cluster sizes.
func CollapseInto(a, b *gene.CodingTranscript) *gene.CodingTranscript {

	descA, clSizeA := parseDesc(a.Desc)
	descB, clSizeB := parseDesc(b.Desc)

	if descA != descB {
		panic("Locus ID mismatch!")
	}

	mergedSize := clSizeA + clSizeB

	a.Desc = descA + "\n" + fmt.Sprintf("%d", mergedSize)

	return a
}

// Attempt to merge together two transcripts in the forward direction.
func MergeForward(a, b *gene.CodingTranscript, fiveTol int, intTol int) (*gene.CodingTranscript, bool) {

	exonsA := a.Exons()
	exonsB := b.Exons()

	// Figure out which transcript has more exons:
	longE, long := exonsA, a
	shortE, short := exonsB, b

	if len(exonsB) > len(exonsA) {
		longE, long = exonsB, b
		shortE, short = exonsA, a
	}

	// Starting from the 3' end check wether exons match with each other:
	for i := len(longE) - 1; i >= (len(longE) - len(shortE)); i-- {
		// Exon position in the shorter transcript:
		s := i - (len(longE) - len(shortE))

		// Exon start and end in the longer transcript:
		ExonStartL := long.Location().Start() + long.Start() + longE[i].Start()
		ExonEndL := long.Location().Start() + long.Start() + longE[i].End()

		// Exon start and end in the shorter transcript:
		ExonStartS := short.Location().Start() + short.Start() + shortE[s].Start()
		ExonEndS := short.Location().Start() + short.Start() + shortE[s].End()

		// Calculate differences in start and end positions:
		deltaThree := Abs(ExonEndS - ExonEndL)
		deltaFive := Abs(ExonStartS - ExonStartL)

		// Apply penalties and decide about merging the transcripts:

		if i == len(longE)-1 {
			if len(shortE) > 1 {
				// Last exon of multiexonic transcript, apply internal tolerance at 5' end:
				if deltaFive > intTol {
					return nil, false
				}
			} else {
				// Special case: the shorter transcript is monoexonic, apply 5' tolerance:
				if deltaFive > fiveTol {
					return nil, false
				}
			}
			// No penalty for 3' end as we applied it during defining a "locus".

		} else if s == 0 {
			// First exon of shorter transcript, apply five prime tolerance (5') and internal tolerance (3'):
			if deltaFive > fiveTol {
				return nil, false
			}
			if deltaThree > intTol {
				return nil, false
			}

		} else {
			// Internal exon, apply internal tolerance twice:
			if deltaFive > intTol {
				return nil, false
			}
			if deltaThree > intTol {
				return nil, false
			}

		}

	}

	// Merge shorter transcript into longer as
	// all penalties were passed:
	return CollapseInto(long, short), true
}

// Attempt to merge together two transcripts in the reverse direction.
func MergeReverse(a, b *gene.CodingTranscript, fiveTol int, intTol int) (*gene.CodingTranscript, bool) {
	exonsA := a.Exons()
	exonsB := b.Exons()

	// Figure out which transcript has more exons:
	longE, long := exonsA, a
	shortE, short := exonsB, b

	if len(exonsB) > len(exonsA) {
		longE, long = exonsB, b
		shortE, short = exonsA, a
	}

	// From the 3' end, check if exons match:
	for i := 0; i < len(shortE); i++ {

		// Exon start and end in the longer transcript:
		ExonStartL := long.Location().Start() + long.Start() + longE[i].Start()
		ExonEndL := long.Location().Start() + long.Start() + longE[i].End()

		// Exon start and end in the shorter transcript:
		ExonStartS := short.Location().Start() + short.Start() + shortE[i].Start()
		ExonEndS := short.Location().Start() + short.Start() + shortE[i].End()

		// Calculate exon boundary distances:
		deltaFive := Abs(ExonEndS - ExonEndL)
		deltaThree := Abs(ExonStartS - ExonStartL)

		// Apply penalties:
		if i == 0 {
			if len(shortE) > 1 {
				// Last exon of multi-exonic transcript:
				// No penalty for 3' end as we applied it during the definition of a "locus".
				// Apply internal tolerance on 5' end:
				if deltaFive > intTol {
					return nil, false
				}
			} else {
				// Special case: shorter transcript is monoexonic, apply 5' tolerance:
				if deltaFive > fiveTol {
					return nil, false
				}
			}
		} else if i == len(shortE)-1 {
			// First exon in the shorter transcipt.
			// Apply internal tolerance at 3' and five prime tolerance at 5':
			if deltaThree > intTol {
				return nil, false
			}
			if deltaFive > fiveTol {
				return nil, false
			}
		} else {
			// Internal exon, apply internal tolerance:
			if deltaThree > intTol {
				return nil, false
			}
			if deltaFive > intTol {
				return nil, false
			}

		}
	}

	// Merge the shorter transcriopt into the longer one:
	return CollapseInto(long, short), true
}

// Collapse partial transcripts within a locus:
func CollapseLocus(locus Locus, trs []*gene.CodingTranscript, fiveTol int, intTol int) []*gene.CodingTranscript {

	// Output buffer:
	resTrs := make([]*gene.CodingTranscript, 0, len(trs))

	// Until input buffrer is empty:
	for len(trs) > 0 {

		// Shift input transcript:
		tr := trs[0]
		trs = trs[1:]

		merged := false
		// Try to mertge into all transcripts in the output buffer:
		for i, rtr := range resTrs {
			var mtr *gene.CodingTranscript
			var match bool

			// Try to merge in the respective direction:
			switch locus.Orient {
			case feat.Forward:
				mtr, match = MergeForward(rtr, tr, fiveTol, intTol)
			case feat.Reverse:
				mtr, match = MergeReverse(rtr, tr, fiveTol, intTol)
			}

			// Sucessful merge:
			if match {
				resTrs[i] = mtr // Update output transcript.
				merged = true
				break
			}
		}

		// Unsuccessful merge: append to output buffer:
		if !merged {
			resTrs = append(resTrs, tr)
		}

	}

	return resTrs
}

// Apply collapsing of partial transcripts to all 3' loci:
func CollapsePartial(pool LocusPool, fiveTol int, intTol int) {

	for locus, trs := range pool {
		pool[locus] = CollapseLocus(locus, trs, fiveTol, intTol)
	}

}
