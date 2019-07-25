package main

import (
	//"io"
	"log"
	//"os"
	"runtime"
)

func main() {
	L = NewLogger("polish_clusters: ", log.Ltime)

	// Parse command line arguments:
	args := new(CmdArgs)
	args.Parse()

	// Set the maximum number of OS threads to use:
	runtime.GOMAXPROCS(int(args.MaxProcs))

	// Check for required commands:
	CheckDependencies()

	// Load clusters:
	clusters := LoadClusters(args.ClustersTab)
	// Initialise output channel for consensus fasta:
	outChan, flushChan := NewSeqWriterChan(args.ConsOut, "fasta", 100)

	var allReads map[string]*Seq
	if !args.SmallMem {
		// Read all BAM records if not in low memory mode:
		allReads = LoadAllReadsFromBam(args.InputFiles[0], int(args.MaxProcs))
	}

	// For each cluster:
	for clusterId, readIds := range clusters {
		// Passing the coverage criteria:
		if len(readIds) >= int(args.MinCoverage) {
			// Get the reads:
			var reads []*Seq
			if args.SmallMem {
				reads = LoadReadsFromBam(args.InputFiles[0], readIds, int(args.MaxProcs))
			} else {
				reads = getClusterFromReads(readIds, allReads)
			}
			// Polish cluster using minimap2 and racon:
			PolishCluster(clusterId, reads, int(args.PolSize), outChan, args.TempDir, int(args.MaxProcs), args.MinimapParams, args.SpoaParams)
		}
	}

	close(outChan)
	<-flushChan

}

// Get cluster of reads from all reads.
func getClusterFromReads(readIds []string, allReads map[string]*Seq) []*Seq {
	res := make([]*Seq, len(readIds))
	for i, readId := range readIds {
		res[i] = allReads[readId]
	}
	return res
}
