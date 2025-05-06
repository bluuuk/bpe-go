package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var _ BPEProcessor = &TiktokenProcessor{}

type TiktokenProcessor struct {
	tokenToRank map[Token]Rank // Encode
	rankToToken map[Rank]Token // Decode
	//maxTokenLength int

	replacementValueInvalidUTF8 []byte
	specialTokens               map[Token]Rank
	allowedSpecialTokens        map[Token]struct{}
	keepUnknownBytes            bool
}

// Decode implements BPEProcessor.
func (t *TiktokenProcessor) Decode(input []Rank) (string, error) {

	buf := &bytes.Buffer{}
	buf.Grow(len(input))

	for _, r := range input {
		tok, ok := t.rankToToken[r]
		if !ok {
			return "", fmt.Errorf("Invalid rank %d", r)
		}
		buf.WriteString(string(tok))
	}

	if len(t.replacementValueInvalidUTF8) > 0 {
		return string(bytes.ToValidUTF8(buf.Bytes(), t.replacementValueInvalidUTF8)), nil
	}

	return buf.String(), nil
}

// Encode implements BPEProcessor.
func (t *TiktokenProcessor) Encode(input string) ([]Rank, error) {

	if len(input) == 0 {
		return []Rank{}, nil
	}

	if len(input) == 1 {
		return []Rank{t.tokenToRank[Token(input)]}, nil
	}

	cleanSpecialTokens := make([]string, 0, len(t.allowedSpecialTokens))
	for token := range t.specialTokens {
		if _, ok := t.allowedSpecialTokens[token]; !ok {
			input = strings.ReplaceAll(input, string(token), "")
		} else {
			cleanSpecialTokens = append(
				cleanSpecialTokens,
				regexp.QuoteMeta(string(token)),
			)
		}
	}

	specialTokenRegex := ""
	if len(cleanSpecialTokens) > 0 {
		specialTokenRegex = "(" + strings.Join(cleanSpecialTokens, "|") + ")"
	}

	re, err := regexp.Compile(specialTokenRegex)

	if err != nil {
		return nil, fmt.Errorf("Could not create regex for allowed special tokens: %w", err)
	}

	tokens := make([]Token, 0, len(input))
	if specialTokenRegex == "" {
		// No special tokens allowed or found; split character by character
		for i := 0; i < len(input); i++ {
			tokens = append(tokens, Token(input[i:i+1]))
		}
	} else {
		// Split input around allowed special tokens, retaining them
		// re.split DOES NOT WORK HERE AS IT DOES NOT HONOR THE CAP. GROUP
		matches := re.FindAllStringIndex(input, -1)
		lastIndex := 0
		for _, match := range matches {
			start, end := match[0], match[1]
			if lastIndex < start {
				// Add tokens from normal text before special token
				segment := input[lastIndex:start]
				for i := 0; i < len(segment); i++ {
					tokens = append(tokens, Token(segment[i:i+1]))
				}
			}

			tokens = append(tokens, Token(input[match[0]:match[1]]))
			lastIndex = end
		}
		// Handle any trailing normal input after last special token
		if lastIndex < len(input) {
			segment := input[lastIndex:]
			for i := 0; i < len(segment); i++ {
				tokens = append(tokens, Token(segment[i:i+1]))
			}
		}
	}
	/*
		We follow the appraoch of
		<https://github.com/openai/tiktoken/blob/main/tiktoken/_educational.py>
		<https://github.com/karpathy/minbpe/blob/master/exercise.md>

		which aim to always merge the lowest rank
	*/

	for {
		var rank Rank = Rank(math.MaxUint64)
		argrank := -1
		for i := range len(tokens) - 1 {
			if r, ok := t.tokenToRank[tokens[i]+tokens[i+1]]; ok && r < rank {
				argrank = i
				rank = r
			}
		}

		if argrank == -1 {
			break
		}

		tokens[argrank] = tokens[argrank] + tokens[argrank+1]
		// TODO: This is not a really memory friendly way of dealing with this
		tokens = slices.Delete(tokens, argrank+1, argrank+2)
	}

	buf := make([]Rank, len(tokens))
	for i, token := range tokens {
		buf[i] = t.tokenToRank[token]
	}

	return buf, nil
}

// Import implements BPEImporter.
func NewTiktokenProcessor(
	dictionaryFilePath string, replacementValueInvalidUTF8 []byte, keepUnknownBytes bool, specialTokens map[Token]Rank, allowedSpecialTokens []Token,
) (*TiktokenProcessor, error) {

	file, err := os.Open(dictionaryFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	var t *TiktokenProcessor = &TiktokenProcessor{}

	t.replacementValueInvalidUTF8 = replacementValueInvalidUTF8
	t.keepUnknownBytes = keepUnknownBytes
	t.tokenToRank = make(map[Token]Rank)
	t.rankToToken = make(map[Rank]Token)
	t.specialTokens = specialTokens
	t.allowedSpecialTokens = make(map[Token]struct{})
	for _, token := range allowedSpecialTokens {
		if _, ok := t.specialTokens[token]; !ok {
			return nil, fmt.Errorf("Allowed special token %s not part of special tokens", token)
		}
		t.allowedSpecialTokens[token] = struct{}{}
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		data := strings.SplitN(scanner.Text(), " ", 2)

		// Ignore line if there is no delimeter
		if len(data) == 1 {
			continue
		}

		encToken, err := base64.StdEncoding.DecodeString(data[0])
		if err != nil {
			return nil, fmt.Errorf("Invalid base64 data at entry %s (%w)", data, err)
		}

		encRank, err := strconv.ParseUint(data[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Invalid number format at entry %s (%w)", data, err)
		}

		token := Token(encToken)
		rank := Rank(encRank)

		if collisionRank, ok := t.tokenToRank[token]; ok {
			return nil, fmt.Errorf("Duplicate rank entry for %d conflicting %d for token %s", collisionRank, rank, token)
		}

		if collisionToken, ok := t.rankToToken[rank]; ok {
			return nil, fmt.Errorf("Duplicate token entry for %s conflicting %s for rank %d", collisionToken, token, rank)
		}

		t.tokenToRank[token] = rank
		t.rankToToken[rank] = token
	}

	for token, rank := range t.specialTokens {
		if otherToken, ok := t.rankToToken[rank]; ok {
			return nil, fmt.Errorf("There is already %s for the special token %s", otherToken, token)
		}

		if otherRank, ok := t.tokenToRank[Token(token)]; ok {
			return nil, fmt.Errorf("There is already %d for the special token %d", otherRank, rank)
		}

		t.tokenToRank[Token(token)] = rank
		t.rankToToken[rank] = Token(token)
	}

	return t, scanner.Err()
}
