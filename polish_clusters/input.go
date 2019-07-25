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

func readLeftClip(r *sam.Record) int {
	if r.Flags&sam.Unmapped != 0 {
		return 0
	}
	if r.Cigar[0].Type() == sam.CigarSoftClipped || r.Cigar[0].Type() == sam.CigarHardClipped {
		return int(r.Cigar[0].Len())
	}
	return 0
}

func readRightClip(r *sam.Record) int {
	if r.Flags&sam.Unmapped != 0 {
		return 0
	}
	last := len(r.Cigar) - 1
	if r.Cigar[last].Type() == sam.CigarSoftClipped || r.Cigar[last].Type() == sam.CigarHardClipped {
		return int(r.Cigar[last].Len())
	}
	return 0
}

func readAcc(r *sam.Record) float64 {
	var mismatch int
	aux, ok := r.Tag([]byte("NM"))
	if !ok {
		panic("no NM tag")
	}
	var mm int
	var ins int
	var del int
	var skip int
	switch aux.Value().(type) {
	case int:
		mismatch = int(aux.Value().(int))
	case int8:
		mismatch = int(aux.Value().(int8))
	case int16:
		mismatch = int(aux.Value().(int16))
	case int32:
		mismatch = int(aux.Value().(int32))
	case int64:
		mismatch = int(aux.Value().(int64))
	case uint:
		mismatch = int(aux.Value().(uint))
	case uint8:
		mismatch = int(aux.Value().(uint8))
	case uint16:
		mismatch = int(aux.Value().(uint16))
	case uint32:
		mismatch = int(aux.Value().(uint32))
	case uint64:
		mismatch = int(aux.Value().(uint64))
	default:
		panic("Could not parse NM tag: " + aux.String())
	}
	for _, op := range r.Cigar {
		switch op.Type() {
		case sam.CigarMatch, sam.CigarEqual, sam.CigarMismatch:
			mm += op.Len()
		case sam.CigarInsertion:
			ins += op.Len()
		case sam.CigarDeletion:
			del += op.Len()
		case sam.CigarSkipped:
			skip += op.Len()
		default:
			//fmt.Println(op)
		}
	}
	return (1.0 - float64(mismatch)/float64(mm+ins+del)) * 100
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
				seq.Acc = readAcc(record)
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
			stmp := string(record.Seq.Expand())
			seq.Seq = stmp[readLeftClip(record) : len(stmp)-readRightClip(record)]
			seq.Qual = record.Qual
			seq.Rev = bool(record.Flags&sam.Reverse != 0)
			res[seq.Id] = seq
		}
	}

	return res
}
