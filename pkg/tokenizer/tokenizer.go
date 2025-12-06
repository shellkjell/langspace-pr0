package tokenizer

import (
	"unicode"
)

// TokenType represents the type of a token
type TokenType int

const (
	TokenTypeIdentifier TokenType = iota
	TokenTypeString
	TokenTypeSemicolon
	TokenTypeMultilineString
	TokenTypeComment
	// New token types for block syntax
	TokenTypeLeftBrace    // {
	TokenTypeRightBrace   // }
	TokenTypeLeftBracket  // [
	TokenTypeRightBracket // ]
	TokenTypeLeftParen    // (
	TokenTypeRightParen   // )
	TokenTypeColon        // :
	TokenTypeComma        // ,
	TokenTypeDot          // .
	TokenTypeEquals       // =
	TokenTypeArrow        // =>
	TokenTypeDollar       // $
	TokenTypeNumber       // numeric literals
	TokenTypeBoolean      // true/false
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
	// Note: pool was removed as it was not being utilized.
	// If memory optimization becomes necessary for large inputs,
	// consider implementing string interning or token reuse.
}

// New creates a new Tokenizer instance
func New() *Tokenizer {
	return &Tokenizer{}
}

// Tokenize breaks the input string into tokens
func (t *Tokenizer) Tokenize(input string) []Token {
	var tokens []Token
	line := 1
	column := 1
	i := 0

	for i < len(input) {
		switch {
		case input[i] == '#':
			// Handle single-line comments
			startCol := column
			start := i
			for i < len(input) && input[i] != '\n' {
				i++
				column++
			}
			tokens = append(tokens, Token{
				Type:   TokenTypeComment,
				Value:  input[start:i],
				Line:   line,
				Column: startCol,
			})
			// Note: newline will be processed in next iteration

		case unicode.IsSpace(rune(input[i])):
			if input[i] == '\n' {
				line++
				column = 1
			} else {
				column++
			}
			i++

		case input[i] == '{':
			tokens = append(tokens, Token{
				Type:   TokenTypeLeftBrace,
				Value:  "{",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == '}':
			tokens = append(tokens, Token{
				Type:   TokenTypeRightBrace,
				Value:  "}",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == '[':
			tokens = append(tokens, Token{
				Type:   TokenTypeLeftBracket,
				Value:  "[",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == ']':
			tokens = append(tokens, Token{
				Type:   TokenTypeRightBracket,
				Value:  "]",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == '(':
			tokens = append(tokens, Token{
				Type:   TokenTypeLeftParen,
				Value:  "(",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == ')':
			tokens = append(tokens, Token{
				Type:   TokenTypeRightParen,
				Value:  ")",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == ':':
			tokens = append(tokens, Token{
				Type:   TokenTypeColon,
				Value:  ":",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == ',':
			tokens = append(tokens, Token{
				Type:   TokenTypeComma,
				Value:  ",",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == '.':
			tokens = append(tokens, Token{
				Type:   TokenTypeDot,
				Value:  ".",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == '$':
			tokens = append(tokens, Token{
				Type:   TokenTypeDollar,
				Value:  "$",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case input[i] == '=' && i+1 < len(input) && input[i+1] == '>':
			tokens = append(tokens, Token{
				Type:   TokenTypeArrow,
				Value:  "=>",
				Line:   line,
				Column: column,
			})
			i += 2
			column += 2

		case input[i] == '=':
			tokens = append(tokens, Token{
				Type:   TokenTypeEquals,
				Value:  "=",
				Line:   line,
				Column: column,
			})
			i++
			column++

		case unicode.IsDigit(rune(input[i])) || (input[i] == '-' && i+1 < len(input) && unicode.IsDigit(rune(input[i+1]))):
			start := i
			startCol := column
			if input[i] == '-' {
				i++
				column++
			}
			for i < len(input) && (unicode.IsDigit(rune(input[i])) || input[i] == '.') {
				i++
				column++
			}
			tokens = append(tokens, Token{
				Type:   TokenTypeNumber,
				Value:  input[start:i],
				Line:   line,
				Column: startCol,
			})

		case unicode.IsLetter(rune(input[i])) || input[i] == '_':
			start := i
			startCol := column
			for i < len(input) && (unicode.IsLetter(rune(input[i])) || unicode.IsDigit(rune(input[i])) || input[i] == '_' || input[i] == '-') {
				i++
				column++
			}
			value := input[start:i]
			// Check for boolean literals
			if value == "true" || value == "false" {
				tokens = append(tokens, Token{
					Type:   TokenTypeBoolean,
					Value:  value,
					Line:   line,
					Column: startCol,
				})
			} else {
				tokens = append(tokens, Token{
					Type:   TokenTypeIdentifier,
					Value:  value,
					Line:   line,
					Column: startCol,
				})
			}

		case i+2 < len(input) && input[i] == '`' && input[i+1] == '`' && input[i+2] == '`':
			startCol := column
			startLine := line
			i += 3 // Skip opening triple backticks
			column += 3
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
					Column: startCol,
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
				if input[i] == '\\' && i+1 < len(input) {
					// Skip escaped character
					i += 2
					column += 2
					continue
				}
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
	case TokenTypeComment:
		return "COMMENT"
	case TokenTypeLeftBrace:
		return "LEFT_BRACE"
	case TokenTypeRightBrace:
		return "RIGHT_BRACE"
	case TokenTypeLeftBracket:
		return "LEFT_BRACKET"
	case TokenTypeRightBracket:
		return "RIGHT_BRACKET"
	case TokenTypeLeftParen:
		return "LEFT_PAREN"
	case TokenTypeRightParen:
		return "RIGHT_PAREN"
	case TokenTypeColon:
		return "COLON"
	case TokenTypeComma:
		return "COMMA"
	case TokenTypeDot:
		return "DOT"
	case TokenTypeEquals:
		return "EQUALS"
	case TokenTypeArrow:
		return "ARROW"
	case TokenTypeDollar:
		return "DOLLAR"
	case TokenTypeNumber:
		return "NUMBER"
	case TokenTypeBoolean:
		return "BOOLEAN"
	default:
		return "UNKNOWN"
	}
}
