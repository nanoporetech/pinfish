package main

import (
	//	"fmt"
	"log"
	//	"os"
	"runtime"
)

// This application is a reimplementation of a Python script written by Jared Simpson:
// https://github.com/jts/nanopore-rna-analysis/blob/master/nanopore_transcript_abundance.py

func main() {
	L = NewLogger("spliced_bam2gff: ", log.Ltime)

	// Parse command line arguments:
	args := new(CmdArgs)
	args.Parse()

	// Set the maximum number of OS threads to use:
	runtime.GOMAXPROCS(int(args.MaxProcs))

	// Open new PAF reader:
	pafChan := NewPafReader(args.InputFiles[0])

	// Initialise new pool:
	pool := NewTranscriptPool()

	// Load compatibility from mappings:
	pool.LoadCompatibility(pafChan)

	// Estimate abundances by EM:
	pool.EmEstimate(int(args.NrIter))

	// Get final abundances:
	abundances := pool.Abundances()

	// Print out counts:
	SaveCounts(abundances, len(pool.Compat))

}
