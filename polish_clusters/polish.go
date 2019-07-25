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

type byNiceLen []*Seq

func (s byNiceLen) Len() int {
	return len(s)
}

func (s byNiceLen) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byNiceLen) Less(i, j int) bool {
	if len(s[i].Seq) > len(s[j].Seq) {
		return true
	}
	if len(s[i].Seq) < len(s[j].Seq) {
		return false
	}
	return s[i].Acc > s[j].Acc
}

// Polish cluster using minimap2 and racon.
func PolishCluster(clusterId string, reads []*Seq, polSize int, outChan chan *Seq, tempRoot string, threads int, minimapParams, spoaParams string) {
	// Set up working space:
	wspace, err := ioutil.TempDir(tempRoot, "pinfish_"+clusterId+"_")
	wspace, _ = filepath.Abs(wspace)
	if err != nil {
		L.Fatalf("Failed to create temporary directory: %s\n", err)
	}

	L.Printf("Polishing cluster %s of size %d\n", clusterId, len(reads))

	sort.Sort(byNiceLen(reads))
	refSeq := reads[0]
	readsFq := WriteReads(reads, wspace, polSize)

	// Polish reference using racon:
	cons := filepath.Join(wspace, "consensus.fas")
	BashExec(fmt.Sprintf("spoa %s -r 0 %s > %s", spoaParams, readsFq, cons))
	//BashExec(fmt.Sprintf("echo -n '>' > %s; spoa %s -r 0 %s > %s", cons, spoaParams, readsFq, cons))

	if FileSize(cons) > 0 {
		// We have a consensus:
		consSeq := new(Seq)
		consSeq.Seq = ReadSpoaCons(cons)
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
func WriteReads(reads []*Seq, wspace string, maxReads int) string {
	if maxReads > len(reads) {
		maxReads = len(reads)
	}

	readsFastq := filepath.Join(wspace, "reads.fasta")
	outChan, flushChan := NewSeqWriterChan(readsFastq, "fasta", 200)

	for i, read := range reads {
		if i >= maxReads {
			break
		}
		outChan <- read
		i++
	}
	close(outChan)
	<-flushChan

	return readsFastq
}
