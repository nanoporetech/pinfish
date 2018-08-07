package main

import (
	"flag"
	"fmt"
	"os"
)

var Version, Build string

// Struct to hold command line arguments:
type CmdArgs struct {
	InputFiles        []string
	MaxProcs          int64
	InternalTolerance int64
	ThreeTolerance    int64
	FiveTolerance     int64
	MonoDiscard       bool
	UnorientDiscard   bool
	ProfFile          string
}

// Parse command line arguments using the flag package.
func (a *CmdArgs) Parse() {
	var help, version bool

	// Process simple command line parameters:
	flag.Int64Var(&a.InternalTolerance, "d", 5, "Internal exon boundary tolerance.")
	flag.Int64Var(&a.ThreeTolerance, "e", 30, "Three prime exons boundary tolerance.")
	flag.Int64Var(&a.FiveTolerance, "f", 5000, "Five prime exons boundary tolerance.")
	flag.BoolVar(&a.MonoDiscard, "M", false, "Discard monoexonic transcripts.")
	flag.BoolVar(&a.UnorientDiscard, "U", false, "Discard transcripts which are not oriented.")
	flag.BoolVar(&help, "h", false, "Print out help message.")
	flag.Int64Var(&a.MaxProcs, "t", 4, "Number of cores to use.")
	flag.StringVar(&a.ProfFile, "prof", "", "Write out CPU profiling information.")
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
}
