package main

import (
	"github.com/biogo/biogo/feat/gene"
	"github.com/google/uuid"
)

// Struct to hold a transcript cluster:
type TranscriptCluster struct {
	Transcripts []*gene.CodingTranscript
	ID          string
	GroupID     string
	LocusSize   int
}

func (tc TranscriptCluster) IsoPercent() float64 {
	return (100.0 * float64(len(tc.Transcripts))) / float64(tc.LocusSize)
}

// Cluster transcripts from a sorted source.
func ClusterTranscriptStream(trStream chan *gene.CodingTranscript, BoundaryTolerance int, EndBoundaryTolerance int) chan *TranscriptCluster {
	// Output channel:
	clusterChan := make(chan *TranscriptCluster, 1000)

	// Cache to hold transcript belonging to the same group:
	cache := make([]*gene.CodingTranscript, 0, 1000)
	go func() {
		// Pull transcripts:
		for tr := range trStream {

			// If the transcript belongs to the current group:
			if SoftRelated(tr, cache, EndBoundaryTolerance) {
				// Add to cache:
				cache = append(cache, tr)
			} else {
				// We found the next group, copy cache:
				tmp := make([]*gene.CodingTranscript, len(cache))
				copy(tmp, cache)
				// Process group to generate clusters:
				ProcessCache(tmp, BoundaryTolerance, EndBoundaryTolerance, clusterChan)
				// Add the current transcript to cache as new group:
				cache = cache[:1]
				cache[0] = tr
			}
		}
		// Process last group:
		ProcessCache(cache, BoundaryTolerance, EndBoundaryTolerance, clusterChan)

		close(clusterChan)
	}()

	return clusterChan
}

// Serach for a matching cluster in a slice of clusters:
func searchClusters(tr *gene.CodingTranscript, clusters []*TranscriptCluster, BoundaryTolerance, EndBoundaryTolerance int) int {
	// Empty cluster:
	if len(clusters) == 0 {
		return -1
	}
	// Check each cluster:
	for i, cluster := range clusters {
		// Compare against each transcript in cluster:
		for _, targetTr := range cluster.Transcripts {
			if TranscriptsHardRelated(tr, targetTr, BoundaryTolerance, EndBoundaryTolerance) {
				// Found match!
				return i
			}
		}
	}

	// No match!
	return -1
}

// Create new cluster having the specified group id and a unique id.
func NewCluster(groupID string, locusSize int) *TranscriptCluster {
	newCls := new(TranscriptCluster)
	newCls.GroupID = groupID
	newCls.ID = uuid.New().String()
	newCls.LocusSize = locusSize
	newCls.Transcripts = make([]*gene.CodingTranscript, 0, 1)
	return newCls
}

// Process group into clusters.
func ProcessCache(cache []*gene.CodingTranscript, BoundaryTolerance, EndBoundaryTolerance int, clusterChan chan *TranscriptCluster) {

	// Slice to store clusters:
	clusters := make([]*TranscriptCluster, 0, 100)
	// Generate unique group id:
	groupID := uuid.New().String()
	//L.Println(groupID, len(cache))
	// For all transcript in cache:
	for _, tr := range cache {
		// Search for matching cluster:
		nrCls := searchClusters(tr, clusters, BoundaryTolerance, EndBoundaryTolerance)
		if nrCls < 0 {
			// No match found, create new cluster:
			newCls := NewCluster(groupID, len(cache))
			newCls.Transcripts = append(newCls.Transcripts, tr)
			clusters = append(clusters, newCls)
		} else {
			// Add to matching cluster:
			clusters[nrCls].Transcripts = append(clusters[nrCls].Transcripts, tr)
		}
	}

	// Send out clusters:
	for _, cls := range clusters {
		clusterChan <- cls
	}
}

// Check wether transcript belongs to group:
func SoftRelated(tr *gene.CodingTranscript, cache []*gene.CodingTranscript, EndBoundaryTolerance int) bool {
	// Empty cache, new transcript belong here:
	if len(cache) == 0 {
		return true
	}
	// Search for sof matching transcript in cache:
	for _, targetTr := range cache {
		if TranscriptsSoftRelated(tr, targetTr, EndBoundaryTolerance) {
			return true
		}
	}
	return false
}
