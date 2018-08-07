// Copyright (C) 2017 Botond Sipos, Oxford Nanopore Technologies

package main

import (
	"bufio"
	"os"
)

import (
	"github.com/biogo/biogo/io/seqio"
	"github.com/biogo/biogo/io/seqio/fasta"
	"github.com/biogo/biogo/io/seqio/fastq"
	"github.com/biogo/biogo/seq"
)

//Write out sequence records fed into a channel.
func NewSeqWriterChan(file string, format string, chanCap int) (chan *Seq, chan bool) {
	f, err := os.Create(file)
	if err != nil {
		L.Fatalf("Cannot create file: %s", err.Error())
	}
	buff := bufio.NewWriter(f)
	var w seqio.Writer
	switch format {
	case "fasta":
		w = fasta.NewWriter(buff, 100)
	case "fastq":
		w = fastq.NewWriter(buff)
	}
	seqChan := make(chan *Seq, chanCap)
	flushChan := make(chan bool)
	go func() {
		var bs seq.Sequence
		i := 0
		for s := range seqChan {
			switch format {
			case "fasta":
				bs = SeqToLinear(s)
			case "fastq":
				bs = SeqToQLinear(s)
			}
			w.Write(bs)
			i++
		}
		buff.Flush()
		f.Close()
		flushChan <- true
	}()
	return seqChan, flushChan
}
