package main

import (
	"bufio"
	//	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// From https://github.com/lh3/miniasm/blob/master/PAF.md:
// PAF is a text format describing the approximate mapping positions between two set of sequences.
// PAF is TAB-delimited with each line consisting of the following predefined fields:

// Col	Type	Description
// 1	string	Query sequence name
// 2	int	Query sequence length
// 3	int	Query start (0-based)
// 4	int	Query end (0-based)
// 5	char	Relative strand: "+" or "-"
// 6	string	Target sequence name
// 7	int	Target sequence length
// 8	int	Target start on original strand (0-based)
// 9	int	Target end on original strand (0-based)
// 10	int	Number of residue matches
// 11	int	Alignment block length
// 12	int	Mapping quality (0-255; 255 for missing)

type PafRecord struct {
	QueryName    string
	QueryLength  int
	QueryStart   int
	QueryEnd     int
	Strand       string
	TargetName   string
	TargetLength int
	TargetStart  int
	TargetEnd    int
	NrMatches    int
	AlnLength    int
	MapQuality   int
}

// Open file and panic at error:
func openFile(file string) *os.File {
	f, err := os.Open(file)
	if err != nil {
		L.Fatalf("Could not open file: %s", err.Error())
	}
	return f
}

func strToInt(s string) int {
	res, err := strconv.Atoi(s)
	if err != nil {
		L.Fatalf("Could not convert to integer: %s\n", s)
	}
	return res
}

func NewPafReader(paf string) chan *PafRecord {
	pafChan := make(chan *PafRecord, 5000)

	fh := openFile(paf)
	reader := bufio.NewReader(fh)

	go func() {
		for {
			line, err := reader.ReadString('\n') // Read next line.
			if err == io.EOF {
				break
			} else if err != nil {
				L.Fatalf("Failed to read cluster file %s: %s\n", paf, err)
			}
			line = line[:len(line)-1] // Remove newline
			tmp := strings.Split(line, "\t")

			p := new(PafRecord)

			p.QueryName = tmp[0]
			p.QueryLength = strToInt(tmp[1])
			p.QueryStart = strToInt(tmp[2])
			p.QueryEnd = strToInt(tmp[3])
			p.Strand = tmp[4]
			p.TargetName = tmp[5]
			p.TargetLength = strToInt(tmp[6])
			p.TargetStart = strToInt(tmp[7])
			p.TargetEnd = strToInt(tmp[8])
			p.NrMatches = strToInt(tmp[9])
			p.AlnLength = strToInt(tmp[10])
			p.MapQuality = strToInt(tmp[11])

			pafChan <- p

		}
		close(pafChan)
	}()

	return pafChan
}
