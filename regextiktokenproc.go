package main

import (
	"fmt"
	"regexp"
)

var _ BPEProcessor = &TiktokenProcessor{}

type RegexTiktokenProcessor struct {
	tkp   *TiktokenProcessor
	regex *regexp.Regexp
}

// Decode implements BPEProcessor.
func (t *RegexTiktokenProcessor) Decode(input []Rank) (string, error) {
	return t.tkp.Decode(input)
}

// Encode implements BPEProcessor.
func (t *RegexTiktokenProcessor) Encode(input string) ([]Rank, error) {

	matches := t.regex.FindAllString(input, -1)

	if len(matches) == 0 {
		return t.tkp.Encode(input)
	}

	res := []Rank{}
	for _, match := range matches {
		r, err := t.tkp.Encode(match)

		if err != nil {
			return []Rank{}, err
		}

		res = append(res, r...)
	}
	return res, nil
}

// Import implements BPEImporter.
func NewRegexTiktokenProcessor(
	dictionaryFilePath string, replacementValueInvalidUTF8 []byte, keepUnknownBytes bool, specialTokens map[Token]Rank, allowedSpecialTokens []Token, regex string,
) (*RegexTiktokenProcessor, error) {
	t, err := NewTiktokenProcessor(dictionaryFilePath, replacementValueInvalidUTF8, keepUnknownBytes, specialTokens, allowedSpecialTokens)

	if err != nil {
		return nil, err
	}

	reg, err := regexp.Compile(regex)

	if err != nil {
		return nil, fmt.Errorf("Cannot compile regex")
	}

	return &RegexTiktokenProcessor{
		tkp:   t,
		regex: reg,
	}, nil
}
