package main

import (
	"flag"
	"fmt"
	"os"
)

var Version, Build string

// Struct to hold command line arguments:
type CmdArgs struct {
	Quiet                bool
	InputFiles           []string
	MaxProcs             int64
	BoundaryTolerance    int64
	EndBoundaryTolerance int64
	MinCoverage          int64
	MinIsoPercent        float64
	ClustersOut          string
	ProfFile             string
}

// Parse command line arguments using the flag package.
func (a *CmdArgs) Parse() {
	var help, version bool

	// Process simple command line parameters:
	flag.StringVar(&a.ClustersOut, "a", "", "Write clusters in tabular format in this file.")
	flag.Int64Var(&a.BoundaryTolerance, "d", 10, "Exon boundary tolerance.")
	flag.Int64Var(&a.EndBoundaryTolerance, "e", 30, "Terminal exons boundary tolerance.")
	flag.Int64Var(&a.MinCoverage, "c", 10, "Minimum cluster size.")
	flag.Float64Var(&a.MinIsoPercent, "p", 1.0, "Minimum isoform percentage.")
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
	if len(a.InputFiles) > 1 {
		L.Fatalf("The maximum number of input files is one!\n")
	}

}
