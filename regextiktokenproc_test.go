package main

import (
	"reflect"
	"testing"
)

func TestRegexTiktokenProcessor_EncodeDecode(t *testing.T) {
	proc, err := NewTiktokenProcessor(
		"cl100k_base.tiktoken",
		[]byte{0xEF, 0xBF, 0xBD},
		true,
		map[Token]Rank{},
		[]Token{},
	)
	// assert that there is no errro
	if err != nil {
		t.Errorf("NewTiktokenProcessor() error = %v", err)
		return
	}

	procRegex, err := NewRegexTiktokenProcessor(
		"cl100k_base.tiktoken",
		[]byte{0xEF, 0xBF, 0xBD},
		true,
		map[Token]Rank{},
		[]Token{},
		` ?\p{L}+(?:'(?:s|m|d|ll|re|ve|t|nt))?| ?\p{N}+| ?[^\s\p{L}\p{N}]+`,
	)

	// assert that there is no errro
	if err != nil {
		t.Errorf("NewTiktokenProcessor() error = %v", err)
		return
	}

	tests := []struct {
		name     string
		args     string
		wantDiff bool
	}{
		{
			name:     "No regex split happens",
			args:     "aaaaaaaaaaaaaaaaaaaaaaaaaa",
			wantDiff: false,
		},
		{
			name:     "Regex split happens",
			args:     "Hello you'reverycool! 100% awesome.it's100percent!",
			wantDiff: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run a round of encode and decode

			ranks1, err1 := proc.Encode(tt.args)
			ranks2, err2 := procRegex.Encode(tt.args)

			if err1 != nil {
				t.Errorf("TiktokenProcessor.Encode() error = %v", err)
				return
			}

			if err2 != nil {
				t.Errorf("RegexTiktokenProcessor.Encode() error = %v", err)
				return
			}

			if reflect.DeepEqual(ranks1, ranks2) == tt.wantDiff {
				if tt.wantDiff {
					t.Errorf("Expected difference %v, got %v", ranks1, ranks2)
				} else {
					t.Errorf("Expected no difference %v, got %v", ranks1, ranks2)
				}
				return
			}
		})
	}
}
