package main

import (
	"fmt"
	"github.com/biogo/biogo/io/featio/gff"
	"io"
	"os"
)

// Write a slice of GFF features to a writer.
func writeGFFs(gffWriter *gff.Writer, trFeatures []gff.Feature) {
	for _, feat := range trFeatures {
		_, err := gffWriter.Write(&feat)
		if err != nil {
			L.Fatalf("Failed to write feature %s: %s", feat, err)
		}
	}
}

// Create clusters tabular outout and write header.
func CreateTabOut(tabOut string) io.Writer {
	fh, err := os.Create(tabOut)
	if err != nil {
		L.Fatalf("Could not create clusters tabular output %s: %s", tabOut, err)
	}
	fmt.Fprintf(fh, "Read\tCluster\n")
	return fh
}

// Write cluster to tabular file.
func WriteClusterTab(cluster *TranscriptCluster, clustersTabOut io.Writer) {
	for _, tr := range cluster.Transcripts {
		fmt.Fprintf(clustersTabOut, "%s\t%s\n", tr.ID[1:len(tr.ID)-1], cluster.ID)
	}
}
