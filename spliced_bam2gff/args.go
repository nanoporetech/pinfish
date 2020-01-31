package main

import (
	"flag"
	"fmt"
	"os"
)

// Strand inference behaviour:
const (
	StrandTag     = 0 // Use strand tag (XS or ts for minimap2).
	StrandRead    = 1 // Use the strand of the reads.
	StrandTagRead = 2 // Use tag or read orientation if unavailable.
)

var Version, Build string

// Struct to hold command line arguments:
type CmdArgs struct {
	MinimapInput    bool
	ForceStrand     bool
	TagReadStrand   bool
	StrandBehaviour int
	InputFiles      []string
	MaxProcs        int64
}

// Parse command line arguments using the flag package.
func (a *CmdArgs) Parse() {
	var help, version bool

	// Process simple command line parameters:
	flag.BoolVar(&a.MinimapInput, "M", false, "Input is from minimap2.")
	flag.BoolVar(&a.ForceStrand, "s", false, "Use read strand (from BAM flag) as feature orientation.")
	flag.BoolVar(&a.TagReadStrand, "g", false, "Use strand tag as feature orientation then read strand if not available.")
	flag.Int64Var(&a.MaxProcs, "t", 4, "Number of cores to use.")
	flag.BoolVar(&help, "h", false, "Print out help message.")
	flag.BoolVar(&version, "V", false, "Print out version.")

	flag.Parse()
	// Print usage:
	if help {
		flag.Usage()
		os.Exit(0)
	}
	// Print version:
	if version {
		fmt.Printf("version: %s build: %s\n", Version, Build)
		os.Exit(0)
	}

	// Set input files:
	a.InputFiles = flag.Args()

	//Check parameters:
	if a.ForceStrand && a.TagReadStrand {
		L.Fatalf("The -s and -g flags are mutually exclusive!\n")
	}
	if a.ForceStrand {
		a.StrandBehaviour = StrandRead
	}
	if a.TagReadStrand {
		a.StrandBehaviour = StrandTagRead
	}
}
