package main

import (
	"github.com/biogo/biogo/io/featio/gff"
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
