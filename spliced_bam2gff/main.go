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
	for _, inBam := range args.InputFiles {
		// Convert spliced BAM entries to GFF transcripts:
		SplicedBam2GFF(inBam, os.Stdout, int(args.MaxProcs), args.MinimapInput, args.StrandBehaviour)
	}

}
