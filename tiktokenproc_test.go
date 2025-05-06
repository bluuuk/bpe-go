package main

import (
	"reflect"
	"testing"
)

func TestTiktokenProcessor_EncodeDecode(t *testing.T) {
	proc, err := NewTiktokenProcessor(
		"testdata/cl100k_base.tiktoken",
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

	tests := []struct {
		name    string
		args    string
		want    string
		wantErr bool
	}{
		{
			name: "Hello world english",
			args: "Hello world",
			want: "Hello world",
		},
		{
			name: "Hello world chinese",
			args: "你好世界",
			want: "你好世界",
		},
		{
			name: "Hello world emoji",
			args: "👋😊 🌎🌟",
			want: "👋😊 🌎🌟",
		},
		{
			name: "Hello world long",
			args: "Hello, World — the phrase that launched a million programs, echoed in terminals from dimly lit basements to towering data centers, a humble greeting from curious minds to the vast digital cosmos, signaling the birth of code, logic, and limitless potential in the age of information.",
			want: "Hello, World — the phrase that launched a million programs, echoed in terminals from dimly lit basements to towering data centers, a humble greeting from curious minds to the vast digital cosmos, signaling the birth of code, logic, and limitless potential in the age of information.",
		},
		{
			// taken form https://www.fileformat.info/info/charset/UTF-8/list.htm?start=40000
			name: "Random unicode gibberish",
			args: "൦ሖᏄᬄₕ𑅢󿿽",
			want: "൦ሖᏄᬄₕ𑅢󿿽",
		},
		{
			// taken from https://stackoverflow.com/questions/1301402/example-invalid-utf8-string
			name: "Invalid UTF-8 with valid ascii",
			// ASCII per byte rep: ?     (    ?      (
			args: string([]byte{0xf0, 0x28, 0x8c, 0x28}),
			want: "�(�(",
		},
		{
			// taken from https://stackoverflow.com/questions/1301402/example-invalid-utf8-string
			name: "fully invalid UTF-8",
			args: string([]byte{0xa0, 0xa1}),
			want: "�",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run a round of encode and decode

			ranks, err := proc.Encode(tt.args)

			if err != nil {
				t.Errorf("TiktokenProcessor.Encode() error = %v", err)
				return
			}

			compare, err := proc.Decode(ranks)
			if err != nil {
				t.Errorf("TiktokenProcessor.Decode() error = %v", err)
				return
			}
			if compare != tt.want {
				t.Errorf("Expected %s, got %s", tt.want, compare)
				return
			}
		})
	}
}

func TestTiktokenProcessor_Encode(t *testing.T) {
	proc, err := NewTiktokenProcessor(
		"testdata/cl100k_base.tiktoken",
		[]byte{0xEF, 0xBF, 0xBD},
		true,
		map[Token]Rank{},
		[]Token{},
	)

	if err != nil {
		t.Errorf("NewTiktokenProcessor() error = %v", err)
		return
	}

	tests := []struct {
		name    string
		args    string
		want    []Rank
		wantErr bool
	}{
		{
			name: "CL100k Base",
			args: "Hello World",
			want: []Rank{
				9906, 4435,
			},
			wantErr: false,
		},
		{
			name: "CL100k Base",
			args: "Hello world",
			want: []Rank{
				9906, 1917,
			},
			wantErr: false,
		},
		{
			name:    "CL100k Base Empty",
			args:    "",
			want:    []Rank{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := proc
			got, err := tr.Encode(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("TiktokenProcessor.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TiktokenProcessor.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTiktokenProcessor_EncodeSpecialTokens(t *testing.T) {
	tests := []struct {
		name    string
		field   func() *TiktokenProcessor
		args    string
		want    []Rank
		wantErr bool
	}{
		{
			name: "CL100k non allowed special token",
			field: func() *TiktokenProcessor {
				proc, err := NewTiktokenProcessor(
					"testdata/cl100k_base.tiktoken",
					[]byte{0xEF, 0xBF, 0xBD},
					true,
					map[Token]Rank{"<|SYSTEM|>": Rank(1 << 62)},
					[]Token{},
				)
				if err != nil {
					t.Fatalf("Bad arguments")
				}
				return proc
			},
			args:    "<|SYSTEM|>",
			want:    []Rank{},
			wantErr: false,
		},
		{
			name: "CL100k allowed special token",
			field: func() *TiktokenProcessor {
				proc, err := NewTiktokenProcessor(
					"testdata/cl100k_base.tiktoken",
					[]byte{0xEF, 0xBF, 0xBD},
					true,
					map[Token]Rank{"<|SYSTEM|>": Rank(1 << 62)},
					[]Token{"<|SYSTEM|>"},
				)
				if err != nil {
					t.Fatalf("Bad arguments")
				}
				return proc
			},
			args:    "<|SYSTEM|>",
			want:    []Rank{Rank(1 << 62)},
			wantErr: false,
		},
		{
			name: "CL100k multiple allowed special token with normal tokens",
			field: func() *TiktokenProcessor {
				proc, err := NewTiktokenProcessor(
					"testdata/cl100k_base.tiktoken",
					[]byte{0xEF, 0xBF, 0xBD},
					true,
					map[Token]Rank{
						"<|SYSTEM|>": Rank(1 << 62),
						"<|USER|>":   Rank(2 << 62),
					},
					[]Token{"<|SYSTEM|>", "<|USER|>"},
				)
				if err != nil {
					t.Fatalf("Bad arguments")
				}
				return proc
			},
			args:    "a<|SYSTEM|>b<|USER|>c<|SYSTEM|>d",
			want:    []Rank{Rank(64), Rank(1 << 62), Rank(65), Rank(2 << 62), Rank(66), Rank(1 << 62), Rank(67)},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := tt.field()
			got, err := tr.Encode(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("TiktokenProcessor.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TiktokenProcessor.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTiktokenProcessor_DecodeSpecialTokens(t *testing.T) {
	tests := []struct {
		name  string
		field func() *TiktokenProcessor
		args  []Rank
		want  string
	}{
		{
			name: "CL100k emits special token",
			field: func() *TiktokenProcessor {
				proc, err := NewTiktokenProcessor(
					"testdata/cl100k_base.tiktoken",
					[]byte{0xEF, 0xBF, 0xBD},
					true,
					map[Token]Rank{"<|SYSTEM|>": Rank(1 << 62)},
					[]Token{},
				)
				if err != nil {
					t.Fatalf("Bad arguments")
				}
				return proc
			},
			args: []Rank{Rank(64), Rank(1 << 62), Rank(65)},
			want: "a<|SYSTEM|>b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := tt.field()
			got, err := tr.Decode(tt.args)
			if err != nil || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TiktokenProcessor.Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}
