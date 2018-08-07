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
	L = NewLogger("collapse_partial: ", log.Ltime)

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

	// Load transcripts into 3' loci:
	locusPool := LoadLoci(trsChan, int(args.ThreeTolerance), args.MonoDiscard, args.UnorientDiscard)

	// Collapse partial transcripts into longer ones:
	CollapsePartial(locusPool, int(args.FiveTolerance), int(args.InternalTolerance))

	// Sort transcript by chromosome names and coordinates:
	trsPool := SortTranscripts(FlattenLocusPool(locusPool))

	// Write out transcriopts in GFF2 format:
	for _, tr := range trsPool {
		gff := Transcript2GFF(tr)
		writeGFFs(gffWriter, gff)

	}

}

// Store all transcipts in one slice.
func FlattenLocusPool(locusPool LocusPool) []*gene.CodingTranscript {
	trsPool := make([]*gene.CodingTranscript, 0, 10000)
	for _, trs := range locusPool {
		trsPool = append(trsPool, trs...)
	}
	return trsPool
}
