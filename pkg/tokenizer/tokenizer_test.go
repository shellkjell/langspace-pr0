package tokenizer

import (
	"testing"
)

func TestTokenizer_Tokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "single_file_entity",
			input: `file "test.txt" path;`,
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "file", Line: 1, Column: 1},
				{Type: TokenTypeString, Value: "test.txt", Line: 1, Column: 6},
				{Type: TokenTypeIdentifier, Value: "path", Line: 1, Column: 17},
				{Type: TokenTypeSemicolon, Value: ";", Line: 1, Column: 21},
			},
		},
		{
			name:  "multiple_entities",
			input: `file "test.txt" path; agent "gpt-4" model;`,
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "file", Line: 1, Column: 1},
				{Type: TokenTypeString, Value: "test.txt", Line: 1, Column: 6},
				{Type: TokenTypeIdentifier, Value: "path", Line: 1, Column: 17},
				{Type: TokenTypeSemicolon, Value: ";", Line: 1, Column: 21},
				{Type: TokenTypeIdentifier, Value: "agent", Line: 1, Column: 23},
				{Type: TokenTypeString, Value: "gpt-4", Line: 1, Column: 29},
				{Type: TokenTypeIdentifier, Value: "model", Line: 1, Column: 37},
				{Type: TokenTypeSemicolon, Value: ";", Line: 1, Column: 42},
			},
		},
		{
			name:  "with_whitespace",
			input: `file   "test.txt"    path;    agent "gpt-4" model;`,
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "file", Line: 1, Column: 1},
				{Type: TokenTypeString, Value: "test.txt", Line: 1, Column: 8},
				{Type: TokenTypeIdentifier, Value: "path", Line: 1, Column: 22},
				{Type: TokenTypeSemicolon, Value: ";", Line: 1, Column: 26},
				{Type: TokenTypeIdentifier, Value: "agent", Line: 1, Column: 31},
				{Type: TokenTypeString, Value: "gpt-4", Line: 1, Column: 37},
				{Type: TokenTypeIdentifier, Value: "model", Line: 1, Column: 45},
				{Type: TokenTypeSemicolon, Value: ";", Line: 1, Column: 50},
			},
		},
		{
			name:  "multiline_string",
			input: "file\n\"test.txt\"\npath;",
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "file", Line: 1, Column: 1},
				{Type: TokenTypeString, Value: "test.txt", Line: 2, Column: 1},
				{Type: TokenTypeIdentifier, Value: "path", Line: 3, Column: 1},
				{Type: TokenTypeSemicolon, Value: ";", Line: 3, Column: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := New()
			got := tokenizer.Tokenize(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf("Tokenize() got %d tokens, want %d", len(got), len(tt.expected))
				for i, token := range got {
					t.Logf("Token[%d] = {Type: %v, Value: %q, Line: %d, Column: %d}", i, token.Type, token.Value, token.Line, token.Column)
				}
				return
			}

			for i, token := range got {
				if token.Type != tt.expected[i].Type {
					t.Errorf("Token[%d].Type = %v, want %v", i, token.Type, tt.expected[i].Type)
				}
				if token.Value != tt.expected[i].Value {
					t.Errorf("Token[%d].Value = %q, want %q", i, token.Value, tt.expected[i].Value)
				}
				if token.Line != tt.expected[i].Line {
					t.Errorf("Token[%d].Line = %d, want %d", i, token.Line, tt.expected[i].Line)
				}
				if token.Column != tt.expected[i].Column {
					t.Errorf("Token[%d].Column = %d, want %d", i, token.Column, tt.expected[i].Column)
				}
			}
		})
	}
}

func TestTokenType_String(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		want      string
	}{
		{TokenTypeIdentifier, "IDENTIFIER"},
		{TokenTypeString, "STRING"},
		{TokenTypeSemicolon, "SEMICOLON"},
		{TokenType(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.tokenType.String(); got != tt.want {
				t.Errorf("TokenType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkTokenizer_Tokenize(b *testing.B) {
	input := `file config.json "contents";
agent validator "check(config.json)";
file script.sh "#!/bin/bash
echo 'Hello World'
exit 0";`

	tokenizer := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer.Tokenize(input)
	}
}
