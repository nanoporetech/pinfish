package main

import (
	"github.com/biogo/biogo/feat/gene"
	"github.com/biogo/biogo/io/featio/gff"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

func main() {
	L = NewLogger("sort_gff: ", log.Ltime)

	// Parse command line arguments:
	args := new(CmdArgs)
	args.Parse()

	// Set the maximum number of OS threads to use:
	runtime.GOMAXPROCS(int(args.MaxProcs))

	// Start up CPU profiling:
	if args.ProfFile != "" {
		f, err := os.Create(args.ProfFile)
		if err != nil {
			L.Fatalf("Could not create file \"%s\" for profiling output: %s", args.ProfFile, err.Error())
		}
		pprof.StartCPUProfile(f)
		defer f.Close()
		defer pprof.StopCPUProfile()
	}

	// Create new GFF writer on standard output:
	gffWriter := gff.NewWriter(os.Stdout, 1000, true)

	// Request channel with input transcripts:
	trsChan := ReadTranscripts(args.InputFiles)

	trsPool := make([]*gene.CodingTranscript, 0, 15000)
	// Load all transcripts:
	for tr := range trsChan {
		trsPool = append(trsPool, tr)
	}

	// Sort transcript by chromosome names and coordinates:
	trsPool = SortTranscripts(trsPool)
	// Write out transcriopts in GFF2 format:
	for _, tr := range trsPool {
		gff := Transcript2GFF(tr)
		writeGFFs(gffWriter, gff)

	}

}
