package main

import (
	"flag"
	"fmt"
	"os"
)

var Version, Build string

// Struct to hold command line arguments:
type CmdArgs struct {
	InputFiles     []string
	MaxProcs       int64
	NrIter         int64
	CompFile       string
	MinReadLength  int64
	FullLenMax     int64
	ScoreThreshold float64
	AlnThreshold   float64
	Verbose        bool
}

// Parse command line arguments using the flag package.
func (a *CmdArgs) Parse() {
	var help, version bool

	// Process simple command line parameters:
	flag.StringVar(&a.CompFile, "c", "", "Compatibility file.")
	flag.Int64Var(&a.MaxProcs, "t", 4, "Maximum number of cores to use.")
	flag.Int64Var(&a.NrIter, "n", 10, "Number of EM iterations.")
	flag.Int64Var(&a.MinReadLength, "m", 0, "Minimum read length.")
	flag.Int64Var(&a.FullLenMax, "f", 20, "Maximum distance from start when classifying as full length.")
	flag.Float64Var(&a.AlnThreshold, "a", 0.5, "Minimum aligned fraction for the best hit.")
	flag.Float64Var(&a.ScoreThreshold, "s", 0.95, "Score threshold used when considering equivalent hits.")
	flag.BoolVar(&a.Verbose, "v", false, "Be verbose.")
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
	if len(a.InputFiles) != 1 {
		L.Fatalf("Exactly one input file must be specified!")
	}
}
