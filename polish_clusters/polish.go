package main

import (
	"fmt"
	"gonum.org/v1/gonum/stat"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
)

// Polish cluster using minimap2 and racon.
func PolishCluster(clusterId string, reads []*Seq, outChan chan *Seq, tempRoot string, threads int, minimapParams, raconParams string) {
	// Set up working space:
	wspace, err := ioutil.TempDir(tempRoot, "pinfish_"+clusterId+"_")
	wspace, _ = filepath.Abs(wspace)
	if err != nil {
		L.Fatalf("Failed to create temporary directory: %s\n", err)
	}

	L.Printf("Polishing cluster %s of size %d\n", clusterId, len(reads))

	// Picck a reference sequence from cluster:
	ref, refSeq := CreateReference(clusterId, reads, wspace)
	_ = refSeq
	readsFq := WriteReads(reads, wspace)

	// Align reads using minimap2:
	sam := filepath.Join(wspace, "alignments.sam")
	BashExec(fmt.Sprintf("minimap2 -ax map-ont -t %d -k14 %s %s %s | samtools view -h -F 2304 > %s", threads, minimapParams, ref, readsFq, sam))

	// Generate overlaps using minimap2:
	//paf := filepath.Join(wspace, "overlaps.paf")
	//BashExec(fmt.Sprintf("minimap2 -x ava-ont -k14 %s %s > %s", ref, readsFq, paf))

	// Polish reference using racon:
	cons := filepath.Join(wspace, "consensus.fq")
	BashExec(fmt.Sprintf("racon -t %d -q -1 %s %s %s %s > %s", threads, raconParams, readsFq, sam, ref, cons))

	if FileSize(cons) > 0 {
		// We have a consensus:
		consSeq := ReadFirstSeq(cons)
		consSeq.Id = fmt.Sprintf("%s|%d", clusterId, len(reads))
		// Reference read mapped to the reverse strand:
		if refSeq.Rev {
			consSeq.Seq = RevCompDNA(consSeq.Seq)
		}
		outChan <- consSeq
	} else {
		// No consensus, write reference:
		L.Printf("No consensus from cluster %s, using representative sequence!\n", clusterId)
		// Reference read mapped to the reverse strand:
		if refSeq.Rev {
			refSeq.Seq = RevCompDNA(refSeq.Seq)
		}
		outChan <- refSeq
	}

	// Remove all temporary files:
	err = os.RemoveAll(wspace)
	if err != nil {
		L.Fatalf("Failed to remove temporary directory %s: %s\n", wspace, err)
	}
}

// Create a reference sequence for a cluster.
func CreateReference(id string, reads []*Seq, wspace string) (string, *Seq) {
	// Get (a) longest sequence from the cluster:
	//tmp := getLongest(reads)
	//tmp := getShortest(reads)
	tmp := getMedian(reads)
	// Copy seq structure:
	longest := &tmp
	// Set cluster id as identifier:
	longest.Id = id

	// Write reference to workspace:
	ref := filepath.Join(wspace, "reference.fq")
	outChan, flushChan := NewSeqWriterChan(ref, "fastq", 1)
	outChan <- longest
	close(outChan)
	<-flushChan

	return ref, longest
}

// Get (a) longest sequence from cluster:
func getLongest(reads []*Seq) Seq {
	var longest *Seq
	var maxLen int
	for _, s := range reads {
		if len(s.Seq) > maxLen {
			longest = s
			maxLen = len(s.Seq)
		}
	}
	return *longest
}

// Get (a) shortest sequence from cluster:
func getShortest(reads []*Seq) Seq {
	var shortest *Seq
	var minLen int = math.MaxInt64
	for _, s := range reads {
		if len(s.Seq) < minLen {
			shortest = s
			minLen = len(s.Seq)
		}
	}
	return *shortest
}

// Get a segment with median length.
func getMedian(segments []*Seq) Seq {
	lengths := make([]float64, len(segments))

	for i, seq := range segments {
		lengths[i] = float64(len(seq.Seq))
	}
	sort.Float64s(lengths)
	medianLength := stat.Quantile(0.5, stat.Empirical, lengths, nil)
	i := sort.SearchFloat64s(lengths, medianLength)
	//fmt.Println(lengths[i], len(segments))

	for _, seq := range segments {
		if len(seq.Seq) == int(lengths[i]) {
			return *seq
		}
	}
	return Seq{}
}

// Write all reads in a fastq in workspace:
func WriteReads(reads []*Seq, wspace string) string {

	readsFastq := filepath.Join(wspace, "reads.fq")
	outChan, flushChan := NewSeqWriterChan(readsFastq, "fastq", 500)

	for _, read := range reads {
		outChan <- read
	}
	close(outChan)
	<-flushChan

	return readsFastq
}
