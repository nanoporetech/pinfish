package main

import (
	"bufio"
	"github.com/biogo/biogo/feat/gene"
	"github.com/biogo/biogo/io/featio/gff"
	"io"
	"os"
)

// Create new GFF reader from file.
func NewGFFReader(gffFile string) *gff.Reader {
	fh, err := os.Open(gffFile)
	if err != nil {
		L.Fatalf("Could not open input file %s: %s\n", gffFile, err)
	}

	reader := gff.NewReader(bufio.NewReader(fh))
	return reader
}

// Read transcripts from input files.
func ReadTranscripts(InputFiles []string) chan *gene.CodingTranscript {

	// Output channel:
	relChan := make(chan *gene.CodingTranscript, 1000)

	go func() {
		var gffReader *gff.Reader

		// Create GFF reader from file or Stdin:
		if len(InputFiles) > 0 {
			gffReader = NewGFFReader(InputFiles[0])
		} else {
			gffReader = gff.NewReader(bufio.NewReader(os.Stdin))
		}

		var currTr *gene.CodingTranscript // Current transcript.
		exons := make(gene.Exons, 0)      // Exon cache.

		for {
			// Get next feature:
			feat, err := gffReader.Read()

			if err == io.EOF {
				// Set exons for last transcript:
				err := currTr.SetExons(exons...)
				if err != nil {
					L.Fatal("Failed to set exons for: %s\n", currTr.ID)
				}
				relChan <- currTr
				break

			} else if err != nil {
				// Read error:
				L.Fatalf("Failed to read feature: %s\n", err)
			}

			gffFeat, _ := feat.(*gff.Feature)

			switch gffFeat.Feature {
			case "mRNA":
				// Add exons to current transcript and process it:
				if currTr != nil {
					err := currTr.SetExons(exons...)
					if err != nil {
						L.Fatal("Failed to set exons for: %s\n", currTr.ID)
					}
					// Process transcript:
					relChan <- currTr
				}
				// Update current transcript and empty exon cache:
				currTr = Feat2NewCodingTranscript(gffFeat)
				exons = make(gene.Exons, 0)
			case "exon":
				// Add exon to cache:
				exon := Feat2NewExon(gffFeat, currTr)
				exons = append(exons, exon)
			default:
				continue // Ignore all other feature types.

			}

		}

		close(relChan)
	}()

	return relChan

}
