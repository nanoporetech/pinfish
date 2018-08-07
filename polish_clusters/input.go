package main

import (
	"bufio"
	"github.com/biogo/hts/bam"
	"github.com/biogo/hts/sam"
	"io"
	"os"
	"strings"
)

// Type for holding clusters:
type Clusters map[string][]string

// Load clusters from tab separated file.
func LoadClusters(tabIn string) Clusters {
	fh, err := os.Open(tabIn)
	if err != nil {
		L.Fatalf("Failed to open clusters file %s: %s", tabIn, err)
	}
	reader := bufio.NewReader(fh)

	clusters := make(Clusters)

	reader.ReadString('\n') // Discard header

	for {
		line, err := reader.ReadString('\n') // Read next line.
		if err == io.EOF {
			break
		} else if err != nil {
			L.Fatalf("Failed to read cluster file %s: %s\n", tabIn, err)
		}
		line = line[:len(line)-1] // Remove newline
		tmp := strings.Split(line, "\t")
		readId, clusterId := tmp[0], tmp[1]

		clusters[clusterId] = append(clusters[clusterId], readId)

	}

	return clusters
}

// Create new BAM reader from file.
func NewBamReader(bamFile string, nrProc int) *bam.Reader {
	fh, err := os.Open(bamFile)
	if err != nil {
		L.Fatalf("Could not open input file %s: %s\n", bamFile, err)
	}

	reader, err := bam.NewReader(bufio.NewReader(fh), nrProc)
	if err != nil {
		L.Fatalf("Could not create BAM reader for %s: %s\n", bamFile, err)
	}
	return reader
}

// Check if string is memeber of a slice.
func inSlice(s string, slice []string) bool {
	for _, member := range slice {
		if member == s {
			return true
		}
	}
	return false
}

// Load specified reads from BAM file as a slice of *Seq objects.
func LoadReadsFromBam(bamFile string, readIds []string, nrProc int) []*Seq {
	bamReader := NewBamReader(bamFile, nrProc)
	res := make([]*Seq, 0, len(readIds))
	// Ierate over BAM records:
	for {
		record, err := bamReader.Read() // Read next record

		if err == io.EOF {
			break // End of file.
		}

		// For all mapped reads:
		if record.Flags&sam.Unmapped == 0 {
			// Which are in the cluster:
			if inSlice(record.Name, readIds) {
				seq := new(Seq)
				seq.Id = record.Name
				seq.Seq = string(record.Seq.Expand())
				seq.Qual = record.Qual
				seq.Rev = bool(record.Flags&sam.Reverse != 0)
				res = append(res, seq)
			}
		}
	}

	return res
}

// Load all reads from a bam file in a map with the read ids as keys.
func LoadAllReadsFromBam(bamFile string, nrProc int) map[string]*Seq {
	bamReader := NewBamReader(bamFile, nrProc)
	res := make(map[string]*Seq)
	// Ierate over BAM records:
	for {
		record, err := bamReader.Read()

		if err == io.EOF {
			break
		}

		// For all mapped reads:
		if record.Flags&sam.Unmapped == 0 {
			seq := new(Seq)
			seq.Id = record.Name
			seq.Seq = string(record.Seq.Expand())
			seq.Qual = record.Qual
			seq.Rev = bool(record.Flags&sam.Reverse != 0)
			res[seq.Id] = seq
		}
	}

	return res
}
