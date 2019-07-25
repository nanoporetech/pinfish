package main

import (
	//"io"
	"log"
	//"os"
	"fmt"
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

	readThreads := int(args.MaxProcs)
	if readThreads > 5 {
		readThreads = 5
	}

	var allReads map[string]*Seq
	if !args.SmallMem {
		// Read all BAM records if not in low memory mode:
		allReads = LoadAllReadsFromBam(args.InputFiles[0], readThreads)
	}

	// For each cluster:
	type ClsInfo struct {
		Id    string
		Reads []*Seq
	}
	type ClsPool chan *ClsInfo
	pool := make(ClsPool, int(args.MaxProcs))
	polDoneChan := make(chan string, args.MaxProcs)

	for clusterId, readIds := range clusters {
		if len(readIds) < int(args.MinCoverage) {
			delete(clusters, clusterId)
		}
	}

	go func() {
		for clusterId, readIds := range clusters {
			// Get the reads:
			var reads []*Seq
			if args.SmallMem {
				reads = LoadReadsFromBam(args.InputFiles[0], readIds, readThreads)
			} else {
				reads = getClusterFromReads(readIds, allReads)
			}
			// Add clusters to pool:
			pool <- &ClsInfo{clusterId, reads}

		}
		close(pool)
	}()

	limit := int(args.MaxProcs)
	nrDone := 0
	if len(clusters) < limit {
		limit = len(clusters)
	}
	for i := 0; i < limit; i++ {
		batch := <-pool
		go PolishCluster(batch.Id, batch.Reads, int(args.PolSize), outChan, polDoneChan, args.TempDir, int(args.MaxProcs), args.MinimapParams, args.SpoaParams)
	}
	for batch := range pool {
		L.Printf(<-polDoneChan)
		nrDone++
		go PolishCluster(batch.Id, batch.Reads, int(args.PolSize), outChan, polDoneChan, args.TempDir, int(args.MaxProcs), args.MinimapParams, args.SpoaParams)
	}
	for i := nrDone; i < len(clusters); i++ {
		L.Printf(<-polDoneChan)
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
