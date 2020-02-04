package main

import (
	"log"
	"os"
	"runtime"
)

func main() {
	L = NewLogger("spliced_bam2gff: ", log.Ltime)

	// Parse command line arguments:
	args := new(CmdArgs)
	args.Parse()

	// Set the maximum number of OS threads to use:
	runtime.GOMAXPROCS(int(args.MaxProcs))

	// Iterate over input files:
	if len(args.InputFiles) != 0 {
		for _, inBam := range args.InputFiles {
			bamReader := NewBamReader(inBam, int(args.MaxProcs))
			// Convert spliced BAM entries to GFF transcripts:
			SplicedBam2GFF(bamReader, os.Stdout, int(args.MaxProcs), args.MinimapInput, args.StrandBehaviour)
		}
	} else {
		bamReader := NewSTDINReader(int(args.MaxProcs))
		SplicedBam2GFF(bamReader, os.Stdout, int(args.MaxProcs), args.MinimapInput, args.StrandBehaviour)
	}
}
