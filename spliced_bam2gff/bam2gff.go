package main

import (
	"fmt"
	"io"

	"github.com/biogo/biogo/feat"
	"github.com/biogo/biogo/feat/gene"
	"github.com/biogo/biogo/feat/genome"
	"github.com/biogo/biogo/io/featio/gff"
	"github.com/biogo/biogo/seq"
	"github.com/biogo/hts/bam"
	"github.com/biogo/hts/sam"
)

// Turn a BAM file containing sliced alignments into GFF2 format annotation.
func SplicedBam2GFF(inReader *bam.Reader, out io.Writer, nrProcBam int, minimapInput bool, strandBehaviour int) {

	gffWriter := gff.NewWriter(out, 1000, true)

	// Ierate over BAM records:
	for {
		record, err := inReader.Read()

		if err == io.EOF {
			break
		}

		// Turn mapped SAM records into GFF:
		if record.Flags&sam.Unmapped == 0 {
			SplicedSAM2GFF(record, gffWriter, minimapInput, strandBehaviour)
		}
	}
}

// Create a new gene.CodingTranscript object from SAM reference, position and orientation.
func NewCodingTranscript(chrom *sam.Reference, id string, pos int, strand feat.Orientation) *gene.CodingTranscript {

	// This will allocate a new chromosome for each transcript
	// but withing this application that should be OK:
	ch := &genome.Chromosome{
		Chr:      chrom.Name(),
		Desc:     chrom.Name(),
		Length:   chrom.Len(),
		Features: nil,
	}

	tr := &gene.CodingTranscript{
		ID:       id,
		Loc:      ch,
		Offset:   pos,
		Orient:   strand,
		Desc:     id,
		CDSstart: 0,
		CDSend:   0,
	}

	return tr
}

// Get orientation from transcript strand tag (either XS, or ts for minimap2).
func getTrStrand(rec *sam.Record, minimapInput bool) feat.Orientation {
	var aux sam.Aux
	if minimapInput {
		aux, _ = rec.Tag([]byte("ts"))
	} else {
		aux, _ = rec.Tag([]byte("XS"))
	}

	// We got the tag value:
	if aux != nil {
		// Convert tag value to string:
		strand := string(aux.Value().(uint8))
		// Decide orientation:
		switch strand {
		case "+":
			return feat.Forward
		case "-":
			return feat.Reverse
		case "?":
			return feat.NotOriented
		default:
			L.Fatalf("Unknown orientation string: %s\n", strand)
		}
	} else {
		//L.Printf("Missing strand tag in record: %s\n", rec.Name)
	}

	// Missing tag, feature not oriented:
	return feat.NotOriented
}

// Flip orientation:
func flipOrientation(orient feat.Orientation) feat.Orientation {
	switch orient {
	case feat.Forward:
		return feat.Reverse
	case feat.Reverse:
		return feat.Forward
	case feat.NotOriented:
		return feat.NotOriented
	default:
		L.Fatalf("Unknown orientation: %s", orient)
	}
	return feat.NotOriented
}

// Decide on the feature strand depending on the transcript strand tag and read orientation:
func figureStrand(readStrand, trStrand feat.Orientation, minimapInput bool, strandBehaviour int) feat.Orientation {

	// Use read orientation as feature strand:
	if strandBehaviour == StrandRead {
		return readStrand
	}

	var strand feat.Orientation

	// Strand tag is missing:
	if trStrand == feat.NotOriented {
		switch strandBehaviour {
		case StrandTag:
			strand = feat.NotOriented // Strand tag takes precedence, feature is not oriented.
		case StrandTagRead:
			strand = readStrand // Fallback to read orientation.
		}
		return strand
	}

	// Transript strand tag is present:

	switch minimapInput {
	case true:
		if trStrand == feat.Reverse {
			strand = flipOrientation(readStrand) // Flip orientaton.
		} else {
			strand = readStrand // Use read strand.
		}
	case false:
		// Input is not minimap2, use transcript strand tag as feature orientation.
		strand = trStrand
	}

	return strand
}

// Convert SAM record into GFF2 records. Each read will be represented as a distinct transcript.
func SplicedSAM2GFF(record *sam.Record, gffWriter *gff.Writer, minimapInput bool, strandBehaviour int) {

	//Get read strand:
	var readStrand feat.Orientation = feat.Forward
	if record.Flags&sam.Reverse != 0 {
		readStrand = feat.Reverse
	}

	// Get transcript strand:
	trStrand := getTrStrand(record, minimapInput)

	// Decide feature strand:
	strand := figureStrand(readStrand, trStrand, minimapInput, strandBehaviour)

	transcript := NewCodingTranscript(record.Ref, record.Name, record.Pos, strand)

	exons := make(gene.Exons, 0) // To accumulate exons.

	// First exon starts at record position:
	var currBlockStart int = record.Pos
	var currBlockLen int = 0
	var exonNr int = 0

CIGAR_LOOP: // Iterate over CIGAR:
	for _, cigar := range record.Cigar {
		op := cigar.Type()
		length := cigar.Len()

		switch op {

		// Soft clip, hard clip, or insertion - do not consume reference:
		case sam.CigarSoftClipped, sam.CigarHardClipped, sam.CigarInsertion:
			continue CIGAR_LOOP

			// Match, mismatch or deletion - add to current exon length:
		case sam.CigarMatch, sam.CigarEqual, sam.CigarMismatch, sam.CigarDeletion:
			currBlockLen += length

		// N operation:
		case sam.CigarSkipped:
			exonStart := currBlockStart              // Previous exon starting here.
			exonEnd := currBlockStart + currBlockLen // Previous exon ends here.

			// Create exon object:
			exonId := fmt.Sprintf("exon_%d", exonNr)
			exon := gene.Exon{transcript, exonStart - record.Pos, exonEnd - exonStart, exonId}

			// Discard zero length exons - FIXME: maybe this should not happen.
			if exon.Len() > 0 {
				exons = append(exons, exon) // Register exon.
			}

			currBlockLen = 0                  // Reset exon length counter.
			currBlockStart = exonEnd + length // Next exon starts after the N operation.
			exonNr++

		default:
			L.Fatalf("Unsupported CIGAR operation %s\n in record %s\n", op, record.Name) // FIXME
		}

	}

	// Deal with the last exon:
	exonStart := currBlockStart
	exonEnd := currBlockStart + currBlockLen
	exonId := fmt.Sprintf("exon_%d", exonNr)
	exon := gene.Exon{transcript, exonStart - record.Pos, exonEnd - exonStart, exonId}
	if exon.Len() > 0 {
		exons = append(exons, exon)
	}

	// Add exons to the transcript:
	err := transcript.SetExons(exons...)
	if err != nil {
		L.Fatalf("Could not set exons for %s: %s\n", transcript.ID, err)
	}

	// Convert transcript into GFF2 features:
	trFeatures := Transcript2GFF(transcript)

	// Write GFF features:
	for _, feat := range trFeatures {
		_, err := gffWriter.Write(&feat)
		if err != nil {
			L.Fatalf("Failed to write feature %s: %s", feat, err)
		}
	}
}

// Convert a gene.CodingTranscript object into a slice of GFF features.
func Transcript2GFF(tr *gene.CodingTranscript) []gff.Feature {
	res := make([]gff.Feature, 0, len(tr.Exons())+1)

	trFeat := gff.Feature{
		SeqName:        tr.Location().Name(),
		Source:         "pinfish",
		Feature:        "mRNA",
		FeatStart:      tr.Start(),
		FeatEnd:        tr.End(),
		FeatScore:      nil,
		FeatStrand:     seq.Strand(tr.Orient),
		FeatFrame:      gff.NoFrame,
		FeatAttributes: gff.Attributes{gff.Attribute{Tag: "gene_id", Value: "\"" + tr.ID + "\""}, gff.Attribute{Tag: "transcript_id", Value: "\"" + tr.ID + "\";"}},
	}

	res = append(res, trFeat)

	for _, exon := range tr.Exons() {
		exFeat := gff.Feature{
			SeqName:        tr.Location().Name(),
			Source:         "pinfish",
			Feature:        "exon",
			FeatStart:      tr.Offset + exon.Start(),
			FeatEnd:        tr.Offset + exon.End(),
			FeatScore:      nil,
			FeatStrand:     seq.Strand(tr.Orient),
			FeatFrame:      gff.NoFrame,
			FeatAttributes: gff.Attributes{gff.Attribute{Tag: "transcript_id", Value: "\"" + tr.ID + "\";"}},
		}
		res = append(res, exFeat)

	}

	return res
}
