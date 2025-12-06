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
			name:  "block_syntax",
			input: `agent "test" { model: "gpt-4" }`,
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "agent", Line: 1, Column: 1},
				{Type: TokenTypeString, Value: "test", Line: 1, Column: 7},
				{Type: TokenTypeLeftBrace, Value: "{", Line: 1, Column: 14},
				{Type: TokenTypeIdentifier, Value: "model", Line: 1, Column: 16},
				{Type: TokenTypeColon, Value: ":", Line: 1, Column: 21},
				{Type: TokenTypeString, Value: "gpt-4", Line: 1, Column: 23},
				{Type: TokenTypeRightBrace, Value: "}", Line: 1, Column: 31},
			},
		},
		{
			name:  "array_syntax",
			input: `tools: [read_file, write_file]`,
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "tools", Line: 1, Column: 1},
				{Type: TokenTypeColon, Value: ":", Line: 1, Column: 6},
				{Type: TokenTypeLeftBracket, Value: "[", Line: 1, Column: 8},
				{Type: TokenTypeIdentifier, Value: "read_file", Line: 1, Column: 9},
				{Type: TokenTypeComma, Value: ",", Line: 1, Column: 18},
				{Type: TokenTypeIdentifier, Value: "write_file", Line: 1, Column: 20},
				{Type: TokenTypeRightBracket, Value: "]", Line: 1, Column: 30},
			},
		},
		{
			name:  "reference_syntax",
			input: `use: agent("reviewer")`,
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "use", Line: 1, Column: 1},
				{Type: TokenTypeColon, Value: ":", Line: 1, Column: 4},
				{Type: TokenTypeIdentifier, Value: "agent", Line: 1, Column: 6},
				{Type: TokenTypeLeftParen, Value: "(", Line: 1, Column: 11},
				{Type: TokenTypeString, Value: "reviewer", Line: 1, Column: 12},
				{Type: TokenTypeRightParen, Value: ")", Line: 1, Column: 22},
			},
		},
		{
			name:  "dot_access",
			input: `step("analyze").output`,
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "step", Line: 1, Column: 1},
				{Type: TokenTypeLeftParen, Value: "(", Line: 1, Column: 5},
				{Type: TokenTypeString, Value: "analyze", Line: 1, Column: 6},
				{Type: TokenTypeRightParen, Value: ")", Line: 1, Column: 15},
				{Type: TokenTypeDot, Value: ".", Line: 1, Column: 16},
				{Type: TokenTypeIdentifier, Value: "output", Line: 1, Column: 17},
			},
		},
		{
			name:  "variable_reference",
			input: `input: $input`,
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "input", Line: 1, Column: 1},
				{Type: TokenTypeColon, Value: ":", Line: 1, Column: 6},
				{Type: TokenTypeDollar, Value: "$", Line: 1, Column: 8},
				{Type: TokenTypeIdentifier, Value: "input", Line: 1, Column: 9},
			},
		},
		{
			name:  "number_literal",
			input: `temperature: 0.7`,
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "temperature", Line: 1, Column: 1},
				{Type: TokenTypeColon, Value: ":", Line: 1, Column: 12},
				{Type: TokenTypeNumber, Value: "0.7", Line: 1, Column: 14},
			},
		},
		{
			name:  "boolean_literals",
			input: `enabled: true disabled: false`,
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "enabled", Line: 1, Column: 1},
				{Type: TokenTypeColon, Value: ":", Line: 1, Column: 8},
				{Type: TokenTypeBoolean, Value: "true", Line: 1, Column: 10},
				{Type: TokenTypeIdentifier, Value: "disabled", Line: 1, Column: 15},
				{Type: TokenTypeColon, Value: ":", Line: 1, Column: 23},
				{Type: TokenTypeBoolean, Value: "false", Line: 1, Column: 25},
			},
		},
		{
			name:  "arrow_operator",
			input: `"bug" => step "fix"`,
			expected: []Token{
				{Type: TokenTypeString, Value: "bug", Line: 1, Column: 1},
				{Type: TokenTypeArrow, Value: "=>", Line: 1, Column: 7},
				{Type: TokenTypeIdentifier, Value: "step", Line: 1, Column: 10},
				{Type: TokenTypeString, Value: "fix", Line: 1, Column: 15},
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
			input: "instruction: ```\nYou are helpful.\n```",
			expected: []Token{
				{Type: TokenTypeIdentifier, Value: "instruction", Line: 1, Column: 1},
				{Type: TokenTypeColon, Value: ":", Line: 1, Column: 12},
				{Type: TokenTypeMultilineString, Value: "\nYou are helpful.\n", Line: 1, Column: 14},
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
		{TokenTypeLeftBrace, "LEFT_BRACE"},
		{TokenTypeRightBrace, "RIGHT_BRACE"},
		{TokenTypeLeftBracket, "LEFT_BRACKET"},
		{TokenTypeRightBracket, "RIGHT_BRACKET"},
		{TokenTypeColon, "COLON"},
		{TokenTypeComma, "COMMA"},
		{TokenTypeDot, "DOT"},
		{TokenTypeNumber, "NUMBER"},
		{TokenTypeBoolean, "BOOLEAN"},
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
	input := `agent "reviewer" {
	model: "claude-sonnet-4-20250514"
	temperature: 0.3
	instruction: """
		You are a code reviewer.
	"""
	tools: [read_file, write_file, run_tests]
}`

	tokenizer := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer.Tokenize(input)
	}
}

func TestTokenReuse(t *testing.T) {
	tokenizer := New()

	// First tokenization
	input1 := `file "test1.txt" path;`
	tokens1 := tokenizer.Tokenize(input1)
	if len(tokens1) != 4 {
		t.Errorf("Expected 4 tokens, got %d", len(tokens1))
	}

	// Second tokenization with same tokenizer instance
	input2 := `file "test2.txt" path;`
	tokens2 := tokenizer.Tokenize(input2)
	if len(tokens2) != 4 {
		t.Errorf("Expected 4 tokens, got %d", len(tokens2))
	}

	// Verify tokens are independent across calls
	if tokens1[1].Value == tokens2[1].Value {
		t.Error("Token values should be different between tokenization calls")
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
			}
		})
	}
}
