package main

type Token string
type Rank uint64

type BPEProcessor interface {
	Encode(string) ([]Rank, error)
	Decode([]Rank) (string, error)
}
