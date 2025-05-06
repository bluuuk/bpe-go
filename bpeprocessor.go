package main

// A token represents a single unit
type Token string

// A rank represents the index for a token inside the embedding table
type Rank uint64

type BPEProcessor interface {
	// Encode processes a string into ranks, regardless of its underlaying character set
	Encode(string) ([]Rank, error)
	// Decode reverse the endocing but may emit correct UTF-8 if a placeholder value is configured
	Decode([]Rank) (string, error)
}
