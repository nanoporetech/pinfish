package main

import (
	"github.com/biogo/biogo/feat/gene"
)

// Decide wether the transcript start sites are close enough:
func TranscriptsSoftRelated(a, b *gene.CodingTranscript, EndBoundaryTolerance int) bool {
	// Mismatching chromosomes:
	if a.Loc.Name() != b.Loc.Name() {
		return false
	}
	// Starts distance larger than tolerance:
	if Abs(int(a.Start())-int(b.Start())) > EndBoundaryTolerance {
		return false
	}

	return true
}

// Decide wether two transcripts belong to the same cluster:
func TranscriptsHardRelated(a, b *gene.CodingTranscript, BoundaryTolerance, EndBoundaryTolerance int) bool {

	// Mismatching orientation: // Treat here not oriented as match?
	if a.Orientation() != b.Orientation() {
		//L.Println("transcript orientation")
		return false
	}

	// Starts distance larger than tolerance:
	if Abs(int(a.Start())-int(b.Start())) > EndBoundaryTolerance {
		return false
	}

	exonsA := a.Exons()
	exonsB := b.Exons()

	// Mismatched exon numbers:
	if len(exonsA) != len(exonsB) {
		return false
	}

	// Exon-by-exon comparison:
	for i, ax := range exonsA {
		bx := exonsB[i]

		deltaStart := Abs((int(a.Start()) + ax.Start()) - (int(b.Start()) + bx.Start())) // exon starts distance
		deltaEnd := Abs((int(a.Start()) + ax.End()) - (int(b.Start()) + bx.End()))       // exon ends distance

		if i == 0 {
			// Starts of first exon, use EndBoundaryTolerance:
			if deltaStart > (EndBoundaryTolerance) {
				return false
			}

		} else {
			if deltaStart > BoundaryTolerance {
				return false
			}
		}

		if i == (len(exonsA) - 1) {
			// End of last exon, use EndBoundaryTolerance:
			if deltaEnd > (EndBoundaryTolerance) {
				return false
			}
		} else {
			if deltaEnd > BoundaryTolerance {
				return false
			}
		}

	}

	// All criteria passed, transcript are related:
	return true
}
