package main

import (
	"github.com/biogo/biogo/feat"
	"github.com/biogo/biogo/feat/gene"
	"github.com/biogo/biogo/feat/genome"
	"github.com/biogo/biogo/io/featio/gff"
	"github.com/biogo/biogo/seq"
	"strconv"
	"strings"
)

// Convert GFF feature into gene.CodingTranscript
func Feat2NewCodingTranscript(feature *gff.Feature) *gene.CodingTranscript {

	ch := &genome.Chromosome{
		Chr:      feature.SeqName,
		Desc:     feature.SeqName,
		Length:   0,
		Features: nil,
	}

	id := feature.FeatAttributes.Get("transcript_id")

	tr := &gene.CodingTranscript{
		ID:       id,
		Loc:      ch,
		Offset:   feature.FeatStart,
		Orient:   feat.Orientation(feature.FeatStrand),
		Desc:     id,
		CDSstart: 0,
		CDSend:   0,
	}

	return tr
}

// Convert GFF feature to a gene.Exon object
func Feat2NewExon(feature *gff.Feature, tr *gene.CodingTranscript) gene.Exon {

	exonTrId := feature.FeatAttributes.Get("transcript_id")
	// Check for transcript/exon mismatch:
	if exonTrId != tr.Name() {
		L.Fatalf("Exon/Transcript mismatch! Exon transcript id: %s Transcript id: %s\n", exonTrId, tr.Name())
	}
	exonId := feature.FeatAttributes.Get("exon_id")

	exon := gene.Exon{
		Transcript: tr,
		Offset:     feature.FeatStart - tr.Start(),
		Length:     feature.FeatEnd - feature.FeatStart,
		Desc:       exonId,
	}

	return exon
}

// Convert a gene.CodingTranscript object into a slice of gff.Feature objects.
func Transcript2GFF(tr *gene.CodingTranscript) []gff.Feature {
	res := make([]gff.Feature, 0, len(tr.Exons())+1)

	// Extract gene ID and cluster size from description:
	descTmp := strings.Split(tr.Desc, "\n")
	desc, clStr := descTmp[0], descTmp[1]
	clSize, _ := strconv.Atoi(clStr)
	// The cluster size will serve as feature score:
	clSizeF := float64(clSize)

	// Create feature for transcript:
	trFeat := gff.Feature{
		SeqName:        tr.Location().Name(),
		Source:         "pinfish",
		Feature:        "mRNA",
		FeatStart:      tr.Start(),
		FeatEnd:        tr.End(),
		FeatScore:      &clSizeF,
		FeatStrand:     seq.Strand(tr.Orient),
		FeatFrame:      gff.NoFrame,
		FeatAttributes: gff.Attributes{gff.Attribute{Tag: "gene_id", Value: "\"" + desc + "\""}, gff.Attribute{Tag: "transcript_id", Value: "\"" + tr.ID + "\";"}},
	}

	res = append(res, trFeat)

	// Create exon features:
	for _, exon := range tr.Exons() {
		exFeat := gff.Feature{
			SeqName:        tr.Location().Name(),
			Source:         "pinfish",
			Feature:        "exon",
			FeatStart:      tr.Offset + exon.Start(),
			FeatEnd:        tr.Offset + exon.End(),
			FeatScore:      &clSizeF,
			FeatStrand:     seq.Strand(tr.Orient),
			FeatFrame:      gff.NoFrame,
			FeatAttributes: gff.Attributes{gff.Attribute{Tag: "transcript_id", Value: "\"" + tr.ID + "\";"}},
		}
		res = append(res, exFeat)

	}

	return res
}
