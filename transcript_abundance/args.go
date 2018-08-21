package main

import (
	"flag"
	"fmt"
	"os"
)

var Version, Build string

// Struct to hold command line arguments:
type CmdArgs struct {
	InputFiles []string
	MaxProcs   int64
	NrIter     int64
	CompFile   string
	Verbose    bool
}

// Parse command line arguments using the flag package.
func (a *CmdArgs) Parse() {
	var help, version bool

	// Process simple command line parameters:
	flag.StringVar(&a.CompFile, "c", "", "Compatibility file.")
	flag.Int64Var(&a.MaxProcs, "t", 4, "Number of cores to use.")
	flag.Int64Var(&a.NrIter, "n", 10, "Number of EM iterations.")
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
}
