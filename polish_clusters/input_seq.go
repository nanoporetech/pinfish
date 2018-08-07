/*
* Copyright (C) 2017 Botond Sipos, Oxford Nanopore Technologies
 */

package main

import (
	"bufio"
	"os"
)
import (
	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/io/seqio"
	"github.com/biogo/biogo/io/seqio/fasta"
	"github.com/biogo/biogo/io/seqio/fastq"
	"github.com/biogo/biogo/seq/linear"
)

// Open file and panic at error:
func openFile(file string) *os.File {
	f, err := os.Open(file)
	if err != nil {
		L.Fatalf("Could not open file: %s", err.Error())
	}
	return f
}

// Decide between fastq and fasta input formats:
func GuessFormat(file string) string {
	reader := openFile(file)
	sc := bufio.NewScanner(reader)
	sc.Scan()
	first_line := string(sc.Text())
	if string(first_line[0]) == ">" {
		return "fasta"
	} else if string(first_line[0]) == "@" {
		return "fastq"
	} else {
		L.Fatalf("Cannot guess format for file: %s\nFirst line: %s", file, first_line)
	}
	reader.Close()
	return ""
}

// Read sequence records from a file of specified format and feed into a channel.
func NewSeqReader(file string) seqio.Reader {
	format := GuessFormat(file)
	fh := openFile(file)
	var reader seqio.Reader
	switch format {
	case "fasta":
		reader = fasta.NewReader(fh, linear.NewSeq("", nil, alphabet.DNAgapped))
	case "fastq":
		template := linear.NewQSeq("", []alphabet.QLetter{}, alphabet.DNAgapped, alphabet.Sanger)
		reader = fastq.NewReader(fh, template)
	}
	return reader
}

// Read sequence records from a file of specified format and feed into a channel.
func NewSeqReaderF(file string) (seqio.Reader, *os.File) {
	format := GuessFormat(file)
	fh := openFile(file)
	var reader seqio.Reader
	switch format {
	case "fasta":
		reader = fasta.NewReader(fh, linear.NewSeq("", nil, alphabet.DNAgapped))
	case "fastq":
		template := linear.NewQSeq("", []alphabet.QLetter{}, alphabet.DNAgapped, alphabet.Sanger)
		reader = fastq.NewReader(fh, template)
	}
	return reader, fh
}

// Read first sequence from file.
func ReadFirstSeq(file string) *Seq {
	reader, fh := NewSeqReaderF(file)

	seq, err := reader.Read()
	if err != nil {
		L.Fatalf("Failed to read sequence from %s: %s\n", file, err)
	}
	record := new(Seq)
	record.Id = seq.CloneAnnotation().ID
	record.Seq = GetSequence(seq)
	//if format == "fastq" {
	//	record.Qual = GetQualityBytes(seq)
	//}
	fh.Close()
	return record
}
