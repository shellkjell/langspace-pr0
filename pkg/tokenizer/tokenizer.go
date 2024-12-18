package tokenizer

import (
	"unicode"

	"github.com/shellkjell/langspace/internal/pool"
)

// TokenType represents the type of a token
type TokenType int

const (
	TokenTypeIdentifier TokenType = iota
	TokenTypeString
	TokenTypeSemicolon
	TokenTypeMultilineString
)

// Token represents a lexical token
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

// Tokenizer represents a LangSpace tokenizer
type Tokenizer struct {
	pool *pool.TokenPool
}

// New creates a new Tokenizer instance
func New() *Tokenizer {
	return &Tokenizer{
		pool: pool.NewTokenPool(),
	}
}

// Tokenize breaks the input string into tokens
func (t *Tokenizer) Tokenize(input string) []Token {
	var tokens []Token
	line := 1
	column := 1
	i := 0

	for i < len(input) {
		switch {
		case unicode.IsSpace(rune(input[i])):
			if input[i] == '\n' {
				line++
				column = 1
			} else {
				column++
			}
			i++

		case unicode.IsLetter(rune(input[i])):
			start := i
			startCol := column
			for i < len(input) && (unicode.IsLetter(rune(input[i])) || unicode.IsDigit(rune(input[i])) || input[i] == '_' || input[i] == '.') {
				i++
				column++
			}
			tokens = append(tokens, Token{
				Type:   TokenTypeIdentifier,
				Value:  input[start:i],
				Line:   line,
				Column: startCol,
			})

		case i+2 < len(input) && input[i] == '`' && input[i+1] == '`' && input[i+2] == '`':
			startCol := column
			startLine := line
			i += 3 // Skip opening triple backticks
			start := i
			
			// Find closing triple backticks
			for i < len(input) {
				if i+2 < len(input) && input[i] == '`' && input[i+1] == '`' && input[i+2] == '`' {
					break
				}
				if input[i] == '\n' {
					line++
					column = 1
				} else {
					column++
				}
				i++
			}
			
			if i+2 < len(input) {
				tokens = append(tokens, Token{
					Type:   TokenTypeMultilineString,
					Value:  input[start:i],
					Line:   startLine,
					Column: startCol - 2, // Adjust for the space before backticks
				})
				i += 3 // Skip closing triple backticks
				column += 3
			}

		case input[i] == '"':
			startCol := column
			i++ // Skip opening quote
			column++
			start := i
			for i < len(input) && input[i] != '"' {
				if input[i] == '\n' {
					line++
					column = 1
				} else {
					column++
				}
				i++
			}
			if i < len(input) {
				tokens = append(tokens, Token{
					Type:   TokenTypeString,
					Value:  input[start:i],
					Line:   line,
					Column: startCol,
				})
				i++ // Skip closing quote
				column++
			}

		case input[i] == ';':
			tokens = append(tokens, Token{
				Type:   TokenTypeSemicolon,
				Value:  ";",
				Line:   line,
				Column: column,
			})
			i++
			column++

		default:
			// Skip invalid characters
			i++
			column++
		}
	}

	return tokens
}

func (t TokenType) String() string {
	switch t {
	case TokenTypeIdentifier:
		return "IDENTIFIER"
	case TokenTypeString:
		return "STRING"
	case TokenTypeMultilineString:
		return "MULTILINE_STRING"
	case TokenTypeSemicolon:
		return "SEMICOLON"
	default:
		return "UNKNOWN"
	}
}
