package main

import (
	"github.com/biogo/biogo/io/featio/gff"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

func main() {
	L = NewLogger("cluster_gff: ", log.Ltime)

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

	// Create tabular clusters output:
	var clustersTabOut io.Writer
	if args.ClustersOut != "" {
		clustersTabOut = CreateTabOut(args.ClustersOut)
	}

	// Create new GFF writer on standard output:
	gffWriter := gff.NewWriter(os.Stdout, 1000, true)

	// Request channel with input transcripts:
	trsChan := ReadTranscripts(args.InputFiles)
	// Produce clusters of input transcripts:

	clusterChan := ClusterTranscriptStream(trsChan, int(args.BoundaryTolerance), int(args.EndBoundaryTolerance))

	for cluster := range clusterChan {
		// Select clusters with enough coverage:
		if cluster.IsoPercent() >= args.MinIsoPercent && len(cluster.Transcripts) >= int(args.MinCoverage) {
			if clustersTabOut != nil {
				WriteClusterTab(cluster, clustersTabOut)
			}
			// Generate cluster consensus:
			consTr := MedianClusterConsensus(cluster)
			// Write out cluster consensus:
			consGFF := Transcript2GFF(consTr)
			writeGFFs(gffWriter, consGFF)
		}
	}
}
