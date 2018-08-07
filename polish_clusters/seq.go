package main

import (
	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/seq"
	"github.com/biogo/biogo/seq/linear"
)

// Struct to hold sequence with qualities:
type Seq struct {
	Id   string
	Seq  string
	Qual []byte
	Rev  bool
}

// Get sequence as string from a biogo Sequence object.
func GetSequence(s seq.Sequence) string {
	buff := make([]byte, s.Len())
	for i := 0; i < s.Len(); i++ {
		buff[i] = byte(s.At(i).L)
	}
	return string(buff)
}

// Get qualities as slice from a biogo Sequence object.
func GetQualities(s seq.Sequence) []int {
	buff := make([]int, s.Len())
	for i := 0; i < s.Len(); i++ {
		buff[i] = int(s.At(i).Q)
	}
	return buff
}

// Get qualities as slice from a biogo Sequence object.
func GetQualityBytes(s seq.Sequence) []byte {
	buff := make([]byte, s.Len())
	for i := 0; i < s.Len(); i++ {
		buff[i] = byte(s.At(i).Q)
	}
	return buff
}

// Make a new anonymous linear.Seq.
func NewAnonLinearSeq(s string) *linear.Seq {
	return &linear.Seq{Seq: alphabet.BytesToLetters([]byte(s))}
}

// Convert a Seq structure to linear.Seq.
func SeqToLinear(s *Seq) *linear.Seq {
	return linear.NewSeq(s.Id, alphabet.BytesToLetters([]byte(s.Seq)), alphabet.DNA)
}

// Convert a Seq structure to linear.QSeq.
func SeqToQLinear(s *Seq) *linear.QSeq {
	qs := make(alphabet.QLetters, len(s.Seq))
	for i, base := range s.Seq {
		qs[i] = alphabet.QLetter{L: alphabet.Letter(base), Q: alphabet.Qphred(s.Qual[i])}
	}
	return linear.NewQSeq(s.Id, qs, alphabet.DNA, alphabet.Sanger)
}

// Reverse complement DNA sequence.
func RevCompDNA(s string) string {
	size := len(s)
	tmp := make([]byte, size)
	var inBase byte
	var outBase byte
	for i := 0; i < size; i++ {
		inBase = s[i]
		switch inBase {
		case 'A':
			outBase = 'T'
		case 'T':
			outBase = 'A'
		case 'G':
			outBase = 'C'
		case 'C':
			outBase = 'G'
		default:
			outBase = 'N'
		}
		tmp[size-1-i] = outBase
	}
	return string(tmp)
}
