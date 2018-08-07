package main

import (
	"fmt"
	"github.com/biogo/biogo/feat/gene"
	"gonum.org/v1/gonum/stat"
	"sort"
)

// Generate consensus of transcript cluster by taking medians of exon boundaries.
func MedianClusterConsensus(cluster *TranscriptCluster) *gene.CodingTranscript {

	nrExons := len(cluster.Transcripts[0].Exons())

	// Slices to store consensus boundaries:
	consExonStarts := make([]int, nrExons)
	consExonEnds := make([]int, nrExons)

	for i := 0; i < nrExons; i++ {

		// Slices to store boundaries of current exon across
		// all transcripts:
		exonStarts := make([]float64, len(cluster.Transcripts))
		exonEnds := make([]float64, len(cluster.Transcripts))

		// Acumulate boundaries:
		for j, tr := range cluster.Transcripts {
			exonStart := float64(tr.Start()) + float64(tr.Exons()[i].Start())
			exonStarts[j] = exonStart

			exonEnd := float64(tr.Start()) + float64(tr.Exons()[i].End())
			exonEnds[j] = exonEnd
		}

		// Sort boundaries:
		sort.Float64s(exonStarts)
		sort.Float64s(exonEnds)

		// Calculate median boundaries:
		medianStart := stat.Quantile(0.5, stat.Empirical, exonStarts, nil)
		medianEnd := stat.Quantile(0.5, stat.Empirical, exonEnds, nil)

		consExonStarts[i] = int(medianStart)
		consExonEnds[i] = int(medianEnd)

	}

	// Convert consensus boundaries to a gene.CodingTranscript object:
	consTr := ExonStartEndToTranscript(consExonStarts, consExonEnds, cluster.Transcripts[0], cluster.ID, cluster.GroupID, len(cluster.Transcripts))

	return consTr
}

// Convert consensus boundaries to a gene.CodingTranscript object.
func ExonStartEndToTranscript(consExonStarts []int, consExonEnds []int, template *gene.CodingTranscript, id, groupID string, size int) *gene.CodingTranscript {

	// Make a copy of the object template:
	tmp := *template
	consTr := &tmp

	consTr.ID = id
	// Store group ID and cluster size in description:
	consTr.Desc = groupID + fmt.Sprintf("\n%d", size)

	// Transcript starts at start of first exon:
	pos := consExonStarts[0]
	consTr.Offset = pos

	exons := make(gene.Exons, len(consExonStarts))

	// Create exon objects:
	for i, start := range consExonStarts {
		end := consExonEnds[i]
		relStart, relEnd := start-pos, end-pos
		exons[i] = gene.Exon{
			Transcript: consTr,
			Offset:     relStart,
			Length:     relEnd - relStart,
			Desc:       fmt.Sprintf("exon_%d", i),
		}
	}

	err := consTr.SetExons(exons...)
	if err != nil {
		L.Fatalf("Could not set exons for consensus transcript.")
	}

	return consTr
}
