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
		{
			name:  "triple_backtick_string",
			input: "file \"script.sh\" ```\n#!/bin/bash\necho 'Hello World'\nexit 0\n``` contents;",
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "file", Line: 1, Column: 1},
				{Type: TokenTypeString, Value: "script.sh", Line: 1, Column: 6},
				{Type: TokenTypeMultilineString, Value: "\n#!/bin/bash\necho 'Hello World'\nexit 0\n", Line: 1, Column: 16},
				{Type: TokenTypeIdentifier, Value: "contents", Line: 5, Column: 5},
				{Type: TokenTypeSemicolon, Value: ";", Line: 5, Column: 13},
			},
		},
		{
			name:  "triple_backtick_with_embedded_quotes",
			input: "file \"code.go\" ```\nfunc main() {\n    fmt.Println(\"Hello `world`\")\n}\n``` contents;",
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "file", Line: 1, Column: 1},
				{Type: TokenTypeString, Value: "code.go", Line: 1, Column: 6},
				{Type: TokenTypeMultilineString, Value: "\nfunc main() {\n    fmt.Println(\"Hello `world`\")\n}\n", Line: 1, Column: 14},
				{Type: TokenTypeIdentifier, Value: "contents", Line: 5, Column: 5},
				{Type: TokenTypeSemicolon, Value: ";", Line: 5, Column: 13},
			},
		},
		{
			name:     "empty_input",
			input:    "",
			expected: []Token{},
		},
		{
			name:     "only_whitespace",
			input:    "   \n\t  \n  ",
			expected: []Token{},
		},
		{
			name:  "unclosed_string",
			input: `file "unclosed`,
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "file", Line: 1, Column: 1},
			},
		},
		{
			name:  "unclosed_multiline",
			input: "file \"script.sh\" ```\n#!/bin/bash\necho 'test'",
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "file", Line: 1, Column: 1},
				{Type: TokenTypeString, Value: "script.sh", Line: 1, Column: 6},
			},
		},
		{
			name:  "complex_multiline_with_spaces",
			input: "file \"test.txt\" ```\n  indented line\n\tTabbed line\n    Mixed   spaces\n``` path;",
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "file", Line: 1, Column: 1},
				{Type: TokenTypeString, Value: "test.txt", Line: 1, Column: 6},
				{Type: TokenTypeMultilineString, Value: "\n  indented line\n\tTabbed line\n    Mixed   spaces\n", Line: 1, Column: 15},
				{Type: TokenTypeIdentifier, Value: "path", Line: 5, Column: 5},
				{Type: TokenTypeSemicolon, Value: ";", Line: 5, Column: 9},
			},
		},
		{
			name:  "consecutive_semicolons",
			input: "file \"test.txt\" path;;",
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "file", Line: 1, Column: 1},
				{Type: TokenTypeString, Value: "test.txt", Line: 1, Column: 6},
				{Type: TokenTypeIdentifier, Value: "path", Line: 1, Column: 17},
				{Type: TokenTypeSemicolon, Value: ";", Line: 1, Column: 21},
				{Type: TokenTypeSemicolon, Value: ";", Line: 1, Column: 22},
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
		{TokenTypeMultilineString, "MULTILINE_STRING"},
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

func TestTokenPool(t *testing.T) {
	tokenizer := New()

	// First tokenization should create new tokens
	input1 := `file "test1.txt" path;`
	tokens1 := tokenizer.Tokenize(input1)
	if len(tokens1) != 4 {
		t.Errorf("Expected 4 tokens, got %d", len(tokens1))
	}

	// Second tokenization should reuse token pool
	input2 := `file "test2.txt" path;`
	tokens2 := tokenizer.Tokenize(input2)
	if len(tokens2) != 4 {
		t.Errorf("Expected 4 tokens, got %d", len(tokens2))
	}

	// Verify tokens are independent despite pool reuse
	if tokens1[1].Value == tokens2[1].Value {
		t.Error("Token values should be different despite pool reuse")
	}
}

func TestTokenizerEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int // expected number of tokens
	}{
		{"single_char", "a", 1},
		{"single_quote", "\"", 0},   // Unclosed quotes are ignored
		{"single_backtick", "`", 0}, // Single backtick is ignored
		{"newline_only", "\n", 0},
		{"unicode_spaces", " ", 0}, // Only testing ASCII space for now
		{"mixed_whitespace", "\t\n \r\n", 0},
		{"special_chars", "@$%", 0}, // Special characters should be ignored (excluding # which is now a comment)
	}

	tokenizer := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenizer.Tokenize(tt.input)
			if len(got) != tt.want {
				t.Errorf("Tokenize() got %d tokens, want %d for input %q", len(got), tt.want, tt.input)
				for i, token := range got {
					t.Logf("Token[%d] = {Type: %v, Value: %q, Line: %d, Column: %d}", i, token.Type, token.Value, token.Line, token.Column)
				}
			}
		})
	}
}

func TestTokenizer_Comments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "single_line_comment",
			input: "# This is a comment",
			expected: []Token{
				{Type: TokenTypeComment, Value: "# This is a comment", Line: 1, Column: 1},
			},
		},
		{
			name:  "comment_with_entity",
			input: "# Comment\nfile \"test.txt\" path;",
			expected: []Token{
				{Type: TokenTypeComment, Value: "# Comment", Line: 1, Column: 1},
				{Type: TokenTypeIdentifier, Value: "file", Line: 2, Column: 1},
				{Type: TokenTypeString, Value: "test.txt", Line: 2, Column: 6},
				{Type: TokenTypeIdentifier, Value: "path", Line: 2, Column: 17},
				{Type: TokenTypeSemicolon, Value: ";", Line: 2, Column: 21},
			},
		},
		{
			name:  "entity_with_trailing_comment",
			input: "file \"test.txt\" path; # inline comment",
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "file", Line: 1, Column: 1},
				{Type: TokenTypeString, Value: "test.txt", Line: 1, Column: 6},
				{Type: TokenTypeIdentifier, Value: "path", Line: 1, Column: 17},
				{Type: TokenTypeSemicolon, Value: ";", Line: 1, Column: 21},
				{Type: TokenTypeComment, Value: "# inline comment", Line: 1, Column: 23},
			},
		},
		{
			name: "multiple_comments_and_entities",
			input: `# File declarations
file "config.json" contents;
# Agent declarations
agent "validator" instruction;`,
			expected: []Token{
				{Type: TokenTypeComment, Value: "# File declarations", Line: 1, Column: 1},
				{Type: TokenTypeIdentifier, Value: "file", Line: 2, Column: 1},
				{Type: TokenTypeString, Value: "config.json", Line: 2, Column: 6},
				{Type: TokenTypeIdentifier, Value: "contents", Line: 2, Column: 20},
				{Type: TokenTypeSemicolon, Value: ";", Line: 2, Column: 28},
				{Type: TokenTypeComment, Value: "# Agent declarations", Line: 3, Column: 1},
				{Type: TokenTypeIdentifier, Value: "agent", Line: 4, Column: 1},
				{Type: TokenTypeString, Value: "validator", Line: 4, Column: 7},
				{Type: TokenTypeIdentifier, Value: "instruction", Line: 4, Column: 19},
				{Type: TokenTypeSemicolon, Value: ";", Line: 4, Column: 30},
			},
		},
		{
			name:  "empty_comment",
			input: "#\nfile \"test.txt\" path;",
			expected: []Token{
				{Type: TokenTypeComment, Value: "#", Line: 1, Column: 1},
				{Type: TokenTypeIdentifier, Value: "file", Line: 2, Column: 1},
				{Type: TokenTypeString, Value: "test.txt", Line: 2, Column: 6},
				{Type: TokenTypeIdentifier, Value: "path", Line: 2, Column: 17},
				{Type: TokenTypeSemicolon, Value: ";", Line: 2, Column: 21},
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
