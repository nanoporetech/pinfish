package main

import (
	"flag"
	"fmt"
	"os"
)

var Version, Build string

// Struct to hold command line arguments:
type CmdArgs struct {
	InputFiles    []string
	MaxProcs      int64
	MinCoverage   int64
	PolSize       int64
	ClustersTab   string
	ProfFile      string
	ConsOut       string
	TempDir       string
	MinimapParams string
	SpoaParams    string
	SmallMem      bool
}

// Parse command line arguments using the flag package.
func (a *CmdArgs) Parse() {
	var help, version bool

	// Process simple command line parameters:
	flag.StringVar(&a.ClustersTab, "a", "", "Read cluster memberships in tabular format.")
	flag.StringVar(&a.ConsOut, "o", "", "Output fasta file.")
	flag.Int64Var(&a.MinCoverage, "c", 1, "Minimum cluster size.")
	flag.Int64Var(&a.MaxProcs, "t", 4, "Number of cores to use.")
	flag.Int64Var(&a.PolSize, "S", 150, "Use maximum this many reads for polishing clusters.")
	flag.StringVar(&a.MinimapParams, "x", "", "Arguments passed to minimap2.")
	flag.StringVar(&a.SpoaParams, "y", "", "Arguments passed to racon.")
	flag.StringVar(&a.TempDir, "d", "", "Location of temporary directory.")
	flag.BoolVar(&a.SmallMem, "m", false, "Do not load all reads in memory (slower).")
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
	if len(a.InputFiles) > 1 {
		L.Fatalf("The maximum number of input BAM files is one!\n")
	}
	if len(a.InputFiles) != 1 {
		L.Fatalf("No input BAM file specified!\n")
	}
	if a.ConsOut == "" {
		L.Fatalf("No output fasta file specified!\n")
	}

}
